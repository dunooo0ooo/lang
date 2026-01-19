package sema

import (
	"fmt"

	"github.com/dunooo0ooo/lang/internal/ast"
	"github.com/dunooo0ooo/lang/internal/bytecode"
	"github.com/dunooo0ooo/lang/internal/token"
)

type Checker struct {
	errs []error

	global *Scope
	scope  *Scope

	inFn    bool
	fnRetTy Type

	ExprType map[ast.Expr]Type
}

func New() *Checker {
	g := NewScope(nil)
	return &Checker{
		global:   g,
		scope:    g,
		ExprType: make(map[ast.Expr]Type),
	}
}

func (c *Checker) Errors() []error { return c.errs }

func (c *Checker) Check(prog *ast.Program) {
	for _, it := range prog.Items {
		if fn, ok := it.(*ast.FnDecl); ok {
			c.declareFn(fn)
		}
	}

	for _, it := range prog.Items {
		switch n := it.(type) {
		case *ast.FnDecl:
			c.checkFn(n)
		case *ast.StmtItem:
			c.checkStmt(n.S)
		}
	}
}

func (c *Checker) declareFn(fn *ast.FnDecl) {
	var params []Type
	for _, p := range fn.Params {
		params = append(params, c.typeFromRef(&p.Type))
	}
	ret := T(bytecode.TypeVoid)
	if fn.RetType != nil {
		ret = c.typeFromRef(fn.RetType)
	}

	if !c.global.Declare(Symbol{
		Kind:   SymFn,
		Name:   fn.Name,
		Pos:    fn.FnPos,
		Params: params,
		Ret:    ret,
	}) {
		c.errorf(fn.FnPos, "redeclaration of function %q", fn.Name)
	}
}

func (c *Checker) checkFn(fn *ast.FnDecl) {
	sym, _ := c.global.Lookup(fn.Name)

	oldScope := c.scope
	c.scope = NewScope(c.global)

	oldInFn, oldRet := c.inFn, c.fnRetTy
	c.inFn, c.fnRetTy = true, sym.Ret

	for i, p := range fn.Params {
		pt := sym.Params[i]
		if !c.scope.Declare(Symbol{Kind: SymVar, Name: p.Name, Pos: p.Pos, Ty: pt}) {
			c.errorf(p.Pos, "redeclaration of parameter %q", p.Name)
		}
	}

	c.checkBlock(fn.Body)

	c.inFn, c.fnRetTy = oldInFn, oldRet
	c.scope = oldScope
}

func (c *Checker) checkBlock(b *ast.BlockStmt) {
	old := c.scope
	c.scope = NewScope(old)

	for _, s := range b.Stmts {
		c.checkStmt(s)
	}

	if b.Tail != nil {
		_ = c.checkExpr(b.Tail)
	}

	c.scope = old
}

func (c *Checker) checkStmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.BlockStmt:
		c.checkBlock(n)
	case *ast.LetStmt:
		c.checkLet(n)
	case *ast.AssignStmt:
		c.checkAssign(n)
	case *ast.ReturnStmt:
		c.checkReturn(n)
	case *ast.IfStmt:
		c.checkIfStmt(n)
	case *ast.WhileStmt:
		c.checkWhile(n)
	case *ast.ForStmt:
		c.checkFor(n)
	case *ast.ExprStmt:
		_ = c.checkExpr(n.X)
	default:
		c.errorf(s.Pos(), "unknown stmt")
	}
}

func (c *Checker) checkLet(s *ast.LetStmt) {
	var initTy Type = T(bytecode.TypeVoid)
	if s.Init != nil {
		initTy = c.checkExpr(s.Init)
	}

	declTy := initTy
	if s.Type != nil {
		declTy = c.typeFromRef(s.Type)
	}

	if declTy.Kind == bytecode.TypeVoid {
		c.errorf(s.LetPos, "let %q requires type or initializer", s.Name)
		declTy = T(bytecode.TypeInvalid)
	}

	if s.Init != nil && !c.assignable(declTy, initTy) && initTy.Kind != bytecode.TypeInvalid {
		c.errorf(s.LetPos, "cannot assign %s to %s", initTy, declTy)
	}

	if !c.scope.Declare(Symbol{Kind: SymVar, Name: s.Name, Pos: s.LetPos, Ty: declTy}) {
		c.errorf(s.LetPos, "redeclaration of variable %q", s.Name)
	}
}

func (c *Checker) checkAssign(s *ast.AssignStmt) {
	sym, ok := c.scope.Lookup(s.Name)
	if !ok || sym.Kind != SymVar {
		c.errorf(s.NamePos, "undefined variable %q", s.Name)
		_ = c.checkExpr(s.Value)
		return
	}
	vty := c.checkExpr(s.Value)
	if !c.assignable(sym.Ty, vty) && vty.Kind != bytecode.TypeInvalid {
		c.errorf(s.NamePos, "cannot assign %s to %s", vty, sym.Ty)
	}
}

func (c *Checker) checkReturn(s *ast.ReturnStmt) {
	if !c.inFn {
		c.errorf(s.RetPos, "return outside function")
		if s.Value != nil {
			_ = c.checkExpr(s.Value)
		}
		return
	}

	retTy := T(bytecode.TypeVoid)
	if s.Value != nil {
		retTy = c.checkExpr(s.Value)
	}

	if !c.assignable(c.fnRetTy, retTy) && retTy.Kind != bytecode.TypeInvalid {
		c.errorf(s.RetPos, "return type %s does not match %s", retTy, c.fnRetTy)
	}
}

func (c *Checker) checkIfStmt(s *ast.IfStmt) {
	cty := c.checkExpr(s.Cond)
	if cty.Kind != bytecode.TypeBool && cty.Kind != bytecode.TypeInvalid {
		c.errorf(s.IfPos, "if condition must be bool, got %s", cty)
	}
	c.checkBlock(s.Then)
	if s.Else != nil {
		c.checkStmt(s.Else)
	}
}

func (c *Checker) checkWhile(s *ast.WhileStmt) {
	cty := c.checkExpr(s.Cond)
	if cty.Kind != bytecode.TypeBool && cty.Kind != bytecode.TypeInvalid {
		c.errorf(s.WhilePos, "while condition must be bool, got %s", cty)
	}
	c.checkBlock(s.Body)
}

func (c *Checker) checkFor(s *ast.ForStmt) {
	old := c.scope
	c.scope = NewScope(old)

	if s.Init != nil {
		c.checkStmt(s.Init)
	}
	if s.Cond != nil {
		cty := c.checkExpr(s.Cond)
		if cty.Kind != bytecode.TypeBool && cty.Kind != bytecode.TypeInvalid {
			c.errorf(s.ForPos, "for condition must be bool, got %s", cty)
		}
	}
	if s.Post != nil {
		c.checkStmt(s.Post)
	}
	c.checkBlock(s.Body)

	c.scope = old
}

func (c *Checker) checkExpr(e ast.Expr) Type {
	var ty Type

	switch n := e.(type) {
	case *ast.IntLit:
		ty = T(bytecode.TypeInt)
	case *ast.FloatLit:
		ty = T(bytecode.TypeFloat)
	case *ast.BoolLit:
		ty = T(bytecode.TypeBool)
	case *ast.StringLit:
		ty = T(bytecode.TypeString)
	case *ast.CharLit:
		ty = T(bytecode.TypeChar)
	case *ast.NullLit:
		ty = T(bytecode.TypeNull)

	case *ast.VarRef:
		sym, ok := c.scope.Lookup(n.Name)
		if !ok {
			c.errorf(n.NamePos, "undefined identifier %q", n.Name)
			ty = T(bytecode.TypeInvalid)
		} else if sym.Kind == SymFn {
			ty = T(bytecode.TypeInvalid)
		} else {
			ty = sym.Ty
		}

	case *ast.UnaryExpr:
		ty = c.checkUnary(n)

	case *ast.BinaryExpr:
		ty = c.checkBinary(n)

	case *ast.CallExpr:
		ty = c.checkCall(n)

	case *ast.ArrayLit:
		ty = c.checkArrayLit(n)

	case *ast.IndexExpr:
		ty = c.checkIndex(n)

	case *ast.BlockExpr:
		ty = c.checkBlockExpr(n.Block)

	case *ast.IfExpr:
		ty = c.checkIfExpr(n)

	default:
		c.errorf(e.Pos(), "unknown expr")
		ty = T(bytecode.TypeInvalid)
	}

	c.ExprType[e] = ty
	return ty
}

func (c *Checker) checkBlockExpr(b *ast.BlockStmt) Type {
	old := c.scope
	c.scope = NewScope(old)

	for _, s := range b.Stmts {
		c.checkStmt(s)
	}

	var t Type = T(bytecode.TypeVoid)
	if b.Tail != nil {
		t = c.checkExpr(b.Tail)
	}

	c.scope = old
	return t
}

func (c *Checker) checkIfExpr(e *ast.IfExpr) Type {
	cty := c.checkExpr(e.Cond)
	if cty.Kind != bytecode.TypeBool && cty.Kind != bytecode.TypeInvalid {
		c.errorf(e.IfPos, "if condition must be bool, got %s", cty)
	}

	thenTy := c.checkBlockExpr(e.Then)
	elseTy := c.checkExpr(e.Else)

	if thenTy.Equal(elseTy) {
		return thenTy
	}

	if thenTy.Kind == bytecode.TypeNull && IsRefType(elseTy) {
		return elseTy
	}
	if elseTy.Kind == bytecode.TypeNull && IsRefType(thenTy) {
		return thenTy
	}

	if thenTy.Kind == bytecode.TypeVoid || elseTy.Kind == bytecode.TypeVoid {
		c.errorf(e.IfPos, "if expression branches must return a value (no ';' after last expr)")
		return T(bytecode.TypeInvalid)
	}

	c.errorf(e.IfPos, "if expression branches have different types: %s vs %s", thenTy, elseTy)
	return T(bytecode.TypeInvalid)
}

func (c *Checker) checkUnary(u *ast.UnaryExpr) Type {
	xt := c.checkExpr(u.X)

	switch u.Op {
	case token.MINUS:
		if xt.Kind != bytecode.TypeInt && xt.Kind != bytecode.TypeFloat && xt.Kind != bytecode.TypeInvalid {
			c.errorf(u.OpPos, "unary '-' expects int/float, got %s", xt)
			return T(bytecode.TypeInvalid)
		}
		return xt
	case token.BANG:
		if xt.Kind != bytecode.TypeBool && xt.Kind != bytecode.TypeInvalid {
			c.errorf(u.OpPos, "unary '!' expects bool, got %s", xt)
			return T(bytecode.TypeInvalid)
		}
		return T(bytecode.TypeBool)
	default:
		c.errorf(u.OpPos, "unknown unary op")
		return T(bytecode.TypeInvalid)
	}
}

func (c *Checker) checkBinary(b *ast.BinaryExpr) Type {
	lt := c.checkExpr(b.L)
	rt := c.checkExpr(b.R)

	if lt.Kind == bytecode.TypeInvalid || rt.Kind == bytecode.TypeInvalid {
		return T(bytecode.TypeInvalid)
	}

	switch b.Op {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH:
		if (lt.Kind == bytecode.TypeInt || lt.Kind == bytecode.TypeFloat) && lt.Kind == rt.Kind {
			return lt
		}
		c.errorf(b.OpPos, "arithmetic expects same numeric types, got %s,%s", lt, rt)
		return T(bytecode.TypeInvalid)

	case token.PERCENT:
		if lt.Kind == bytecode.TypeInt && rt.Kind == bytecode.TypeInt {
			return T(bytecode.TypeInt)
		}
		c.errorf(b.OpPos, "%% expects int,int got %s,%s", lt, rt)
		return T(bytecode.TypeInvalid)

	case token.LT, token.LTE, token.GT, token.GTE:
		if (lt.Kind == bytecode.TypeInt || lt.Kind == bytecode.TypeFloat || lt.Kind == bytecode.TypeChar) && lt.Kind == rt.Kind {
			return T(bytecode.TypeBool)
		}
		c.errorf(b.OpPos, "comparison expects same comparable types, got %s,%s", lt, rt)
		return T(bytecode.TypeInvalid)

	case token.EQ, token.NEQ:
		if lt.Equal(rt) {
			return T(bytecode.TypeBool)
		}
		if lt.Kind == bytecode.TypeNull && IsRefType(rt) {
			return T(bytecode.TypeBool)
		}
		if rt.Kind == bytecode.TypeNull && IsRefType(lt) {
			return T(bytecode.TypeBool)
		}
		c.errorf(b.OpPos, "equality expects same types (or null with ref), got %s,%s", lt, rt)
		return T(bytecode.TypeInvalid)

	case token.AND, token.OR:
		if lt.Kind == bytecode.TypeBool && rt.Kind == bytecode.TypeBool {
			return T(bytecode.TypeBool)
		}
		c.errorf(b.OpPos, "logic expects bool,bool got %s,%s", lt, rt)
		return T(bytecode.TypeInvalid)
	}

	c.errorf(b.OpPos, "unknown binary op")
	return T(bytecode.TypeInvalid)
}

func (c *Checker) checkCall(call *ast.CallExpr) Type {
	vr, ok := call.Callee.(*ast.VarRef)
	if !ok {
		c.errorf(call.Pos(), "call target must be identifier")
		for _, a := range call.Args {
			_ = c.checkExpr(a)
		}
		return T(bytecode.TypeInvalid)
	}

	if vr.Name == "println" {
		if len(call.Args) != 1 {
			c.errorf(call.Pos(), "println expects 1 argument")
			for _, a := range call.Args {
				_ = c.checkExpr(a)
			}
			return T(bytecode.TypeVoid)
		}
		_ = c.checkExpr(call.Args[0])
		return T(bytecode.TypeVoid)
	}

	switch vr.Name {
	case "array":
		if len(call.Args) != 1 {
			c.errorf(call.Pos(), "function %q expects %d args, got %d", vr.Name, 1, len(call.Args))
			for _, a := range call.Args {
				_ = c.checkExpr(a)
			}
			return Arr(T(bytecode.TypeInt))
		}
		t0 := c.checkExpr(call.Args[0])
		if t0.Kind != bytecode.TypeInt && t0.Kind != bytecode.TypeInvalid {
			c.errorf(call.Args[0].Pos(), "array(len): len must be int, got %s", t0)
			return T(bytecode.TypeInvalid)
		}
		return Arr(T(bytecode.TypeInt))
	case "print":
		if len(call.Args) != 1 {
			c.errorf(call.Pos(), "function %q expects %d args, got %d", vr.Name, 1, len(call.Args))
			for _, a := range call.Args {
				_ = c.checkExpr(a)
			}
			return T(bytecode.TypeVoid)
		}

		argTy := c.checkExpr(call.Args[0])
		if argTy.Kind == bytecode.TypeVoid {
			c.errorf(call.Args[0].Pos(), "cannot print void")
			return T(bytecode.TypeInvalid)
		}

		return T(bytecode.TypeVoid)

	case "get":
		if len(call.Args) != 2 {
			c.errorf(call.Pos(), "function %q expects %d args, got %d", vr.Name, 2, len(call.Args))
			for _, a := range call.Args {
				_ = c.checkExpr(a)
			}
			return T(bytecode.TypeInt)
		}

		tArr := c.checkExpr(call.Args[0])
		tIdx := c.checkExpr(call.Args[1])

		if !tArr.IsArray() || tArr.Elem == nil || tArr.Elem.Kind != bytecode.TypeInt {
			c.errorf(call.Args[0].Pos(), "get(arr, i): arr must be []int, got %s", tArr)
			return T(bytecode.TypeInvalid)
		}
		if tIdx.Kind != bytecode.TypeInt && tIdx.Kind != bytecode.TypeInvalid {
			c.errorf(call.Args[1].Pos(), "get(arr, i): i must be int, got %s", tIdx)
			return T(bytecode.TypeInvalid)
		}
		return T(bytecode.TypeInt)

	case "set":
		if len(call.Args) != 3 {
			c.errorf(call.Pos(), "function %q expects %d args, got %d", vr.Name, 3, len(call.Args))
			for _, a := range call.Args {
				_ = c.checkExpr(a)
			}
			return T(bytecode.TypeVoid)
		}

		tArr := c.checkExpr(call.Args[0])
		tIdx := c.checkExpr(call.Args[1])
		tVal := c.checkExpr(call.Args[2])

		if !tArr.IsArray() || tArr.Elem == nil || tArr.Elem.Kind != bytecode.TypeInt {
			c.errorf(call.Args[0].Pos(), "set(arr, i, v): arr must be []int, got %s", tArr)
			return T(bytecode.TypeInvalid)
		}
		if tIdx.Kind != bytecode.TypeInt && tIdx.Kind != bytecode.TypeInvalid {
			c.errorf(call.Args[1].Pos(), "set(arr, i, v): i must be int, got %s", tIdx)
			return T(bytecode.TypeInvalid)
		}
		if tVal.Kind != bytecode.TypeInt && tVal.Kind != bytecode.TypeInvalid {
			c.errorf(call.Args[2].Pos(), "set(arr, i, v): v must be int, got %s", tVal)
			return T(bytecode.TypeInvalid)
		}
		return T(bytecode.TypeVoid)
	}

	sym, ok := c.scope.Lookup(vr.Name)
	if !ok || sym.Kind != SymFn {
		c.errorf(vr.NamePos, "undefined function %q", vr.Name)
		for _, a := range call.Args {
			_ = c.checkExpr(a)
		}
		return T(bytecode.TypeInvalid)
	}

	if len(call.Args) != len(sym.Params) {
		c.errorf(call.Pos(), "function %q expects %d args, got %d", vr.Name, len(sym.Params), len(call.Args))
		for _, a := range call.Args {
			_ = c.checkExpr(a)
		}
		return sym.Ret
	}

	for i, a := range call.Args {
		at := c.checkExpr(a)
		pt := sym.Params[i]
		if !c.assignable(pt, at) && at.Kind != bytecode.TypeInvalid {
			c.errorf(a.Pos(), "arg %d: expected %s, got %s", i, pt, at)
		}
	}

	return sym.Ret
}

func (c *Checker) checkArrayLit(a *ast.ArrayLit) Type {
	if len(a.Elems) == 0 {
		c.errorf(a.Lbrack, "empty array literal requires type annotation")
		return Arr(T(bytecode.TypeInvalid))
	}

	elemTy := c.checkExpr(a.Elems[0])
	if elemTy.Kind == bytecode.TypeInvalid {
		return Arr(elemTy)
	}

	for i := 1; i < len(a.Elems); i++ {
		t := c.checkExpr(a.Elems[i])
		if !elemTy.Equal(t) && t.Kind != bytecode.TypeInvalid {
			c.errorf(a.Elems[i].Pos(), "array element %d: expected %s, got %s", i, elemTy, t)
		}
	}

	if elemTy.Kind == bytecode.TypeNull {
		c.errorf(a.Lbrack, "array literal of null requires type annotation")
		return Arr(T(bytecode.TypeInvalid))
	}

	return Arr(elemTy)
}

func (c *Checker) checkIndex(ix *ast.IndexExpr) Type {
	xTy := c.checkExpr(ix.X)
	iTy := c.checkExpr(ix.Index)

	if iTy.Kind != bytecode.TypeInt && iTy.Kind != bytecode.TypeInvalid {
		c.errorf(ix.Index.Pos(), "index must be int, got %s", iTy)
	}

	if xTy.Kind == bytecode.TypeArray {
		if xTy.Elem == nil {
			return T(bytecode.TypeInvalid)
		}
		return *xTy.Elem
	}

	c.errorf(ix.Lbrack, "cannot index %s", xTy)
	return T(bytecode.TypeInvalid)
}

func (c *Checker) assignable(dst, src Type) bool {
	if dst.Equal(src) {
		return true
	}
	if src.Kind == bytecode.TypeNull && IsRefType(dst) {
		return true
	}
	return false
}

func (c *Checker) typeFromRef(r *ast.TypeRef) Type {
	if r == nil {
		return T(bytecode.TypeVoid)
	}

	if r.Name == "array" {
		if r.Elem == nil {
			return Arr(T(bytecode.TypeInvalid))
		}
		elem := c.typeFromRef(r.Elem)
		return Arr(elem)
	}

	switch r.Name {
	case "int":
		return T(bytecode.TypeInt)
	case "float":
		return T(bytecode.TypeFloat)
	case "bool":
		return T(bytecode.TypeBool)
	case "string":
		return T(bytecode.TypeString)
	case "char":
		return T(bytecode.TypeChar)
	case "void":
		return T(bytecode.TypeVoid)
	default:
		return T(bytecode.TypeInvalid)
	}
}

func (c *Checker) errorf(pos token.Position, format string, args ...any) {
	c.errs = append(c.errs, fmt.Errorf("%d:%d: %s", pos.Line, pos.Col, fmt.Sprintf(format, args...)))
}
