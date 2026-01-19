package sema

import (
	"fmt"
	"github.com/dunooo0ooo/lang/internal/bytecode"

	"github.com/dunooo0ooo/lang/internal/ast"
	"github.com/dunooo0ooo/lang/internal/token"
)

type LocalID int
type FuncID int

type ResolvedVar struct {
	ID LocalID
	Ty Type
}

type ResolvedFn struct {
	ID     FuncID
	Params []Type
	Ret    Type
}

type ResolveResult struct {
	Fns  map[*ast.FnDecl]ResolvedFn
	Vars map[ast.Expr]ResolvedVar
	Asgn map[*ast.AssignStmt]ResolvedVar
	Let  map[*ast.LetStmt]ResolvedVar
}

type resolverScope struct {
	parent *resolverScope
	vars   map[string]ResolvedVar
}

func newResolverScope(parent *resolverScope) *resolverScope {
	return &resolverScope{parent: parent, vars: make(map[string]ResolvedVar)}
}

func (s *resolverScope) lookup(name string) (ResolvedVar, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if v, ok := sc.vars[name]; ok {
			return v, true
		}
	}
	return ResolvedVar{}, false
}

func (s *resolverScope) declare(name string, v ResolvedVar) bool {
	if _, ok := s.vars[name]; ok {
		return false
	}
	s.vars[name] = v
	return true
}

type Resolver struct {
	errs []error

	fnIDs map[string]FuncID
	nextF FuncID

	scope *resolverScope
	nextL LocalID

	out ResolveResult

	types map[ast.Expr]Type
}

func NewResolver(exprTypes map[ast.Expr]Type) *Resolver {
	r := &Resolver{
		fnIDs: make(map[string]FuncID),
		out: ResolveResult{
			Fns:  make(map[*ast.FnDecl]ResolvedFn),
			Vars: make(map[ast.Expr]ResolvedVar),
			Asgn: make(map[*ast.AssignStmt]ResolvedVar),
			Let:  make(map[*ast.LetStmt]ResolvedVar),
		},
		types: exprTypes,
	}
	r.scope = newResolverScope(nil)
	return r
}

func (r *Resolver) Errors() []error { return r.errs }
func (r *Resolver) Result() ResolveResult {
	return r.out
}

func (r *Resolver) Resolve(prog *ast.Program) {
	for _, it := range prog.Items {
		if fn, ok := it.(*ast.FnDecl); ok {
			if _, exists := r.fnIDs[fn.Name]; exists {
				continue
			}
			r.fnIDs[fn.Name] = r.nextF
			r.nextF++
		}
	}

	for _, it := range prog.Items {
		switch n := it.(type) {
		case *ast.FnDecl:
			r.resolveFn(n)
		case *ast.StmtItem:
			r.resolveStmt(n.S)
		}
	}
}

func (r *Resolver) resolveFn(fn *ast.FnDecl) {
	fnID := r.fnIDs[fn.Name]

	oldScope := r.scope
	oldNext := r.nextL
	r.scope = newResolverScope(nil)
	r.nextL = 0

	params := make([]Type, 0, len(fn.Params))
	for _, p := range fn.Params {
		pty := r.typeFromParam(p)
		id := r.allocLocal(p.Name, p.Pos, pty)
		_ = id
		params = append(params, pty)
	}

	ret := r.typeFromFnRet(fn)
	r.out.Fns[fn] = ResolvedFn{ID: fnID, Params: params, Ret: ret}

	r.resolveBlock(fn.Body)

	r.scope = oldScope
	r.nextL = oldNext
}

func (r *Resolver) resolveBlock(b *ast.BlockStmt) {
	old := r.scope
	r.scope = newResolverScope(old)
	for _, s := range b.Stmts {
		r.resolveStmt(s)
	}
	r.scope = old
}

func (r *Resolver) resolveStmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.BlockStmt:
		r.resolveBlock(n)
	case *ast.LetStmt:
		r.resolveLet(n)
	case *ast.AssignStmt:
		r.resolveAssign(n)
	case *ast.ReturnStmt:
		if n.Value != nil {
			r.resolveExpr(n.Value)
		}
	case *ast.IfStmt:
		r.resolveExpr(n.Cond)
		r.resolveBlock(n.Then)
		if n.Else != nil {
			r.resolveStmt(n.Else)
		}
	case *ast.WhileStmt:
		r.resolveExpr(n.Cond)
		r.resolveBlock(n.Body)
	case *ast.ForStmt:
		old := r.scope
		r.scope = newResolverScope(old)
		if n.Init != nil {
			r.resolveStmt(n.Init)
		}
		if n.Cond != nil {
			r.resolveExpr(n.Cond)
		}
		if n.Post != nil {
			r.resolveStmt(n.Post)
		}
		r.resolveBlock(n.Body)
		r.scope = old
	case *ast.ExprStmt:
		r.resolveExpr(n.X)
	}
}

func (r *Resolver) resolveLet(s *ast.LetStmt) {
	if s.Init != nil {
		r.resolveExpr(s.Init)
	}
	ty := r.typeFromLet(s)
	id := r.allocLocal(s.Name, s.LetPos, ty)
	r.out.Let[s] = ResolvedVar{ID: id, Ty: ty}
}

func (r *Resolver) resolveAssign(s *ast.AssignStmt) {
	r.resolveExpr(s.Value)
	v, ok := r.scope.lookup(s.Name)
	if !ok {
		r.errorf(s.NamePos, "unresolved variable %q", s.Name)
		return
	}
	r.out.Asgn[s] = v
}

func (r *Resolver) resolveExpr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.VarRef:
		v, ok := r.scope.lookup(n.Name)
		if !ok {
			r.errorf(n.NamePos, "unresolved identifier %q", n.Name)
			return
		}
		r.out.Vars[e] = v

	case *ast.UnaryExpr:
		r.resolveExpr(n.X)
	case *ast.BinaryExpr:
		r.resolveExpr(n.L)
		r.resolveExpr(n.R)
	case *ast.CallExpr:
		r.resolveExpr(n.Callee)
		for _, a := range n.Args {
			r.resolveExpr(a)
		}
	case *ast.ArrayLit:
		for _, el := range n.Elems {
			r.resolveExpr(el)
		}
	case *ast.IndexExpr:
		r.resolveExpr(n.X)
		r.resolveExpr(n.Index)
	case *ast.BlockExpr:
		r.resolveBlock(n.Block)
	case *ast.IfExpr:
		r.resolveExpr(n.Cond)
		r.resolveBlock(n.Then)
		r.resolveExpr(n.Else)
	}
}

func (r *Resolver) allocLocal(name string, pos token.Position, ty Type) LocalID {
	id := r.nextL
	r.nextL++

	if ok := r.scope.declare(name, ResolvedVar{ID: id, Ty: ty}); !ok {
		r.errorf(pos, "redeclaration of %q", name)
	}
	return id
}

func (r *Resolver) typeFromLet(s *ast.LetStmt) Type {
	if s.Type != nil {
		return typeFromRef(s.Type)
	}
	if s.Init != nil && r.types != nil {
		if t, ok := r.types[s.Init]; ok {
			return t
		}
	}
	return T(0)
}

func (r *Resolver) typeFromParam(p ast.Param) Type {
	return typeFromRef(&p.Type)
}

func (r *Resolver) typeFromFnRet(fn *ast.FnDecl) Type {
	if fn.RetType == nil {
		return T(bytecode.TypeVoid)
	}
	return typeFromRef(fn.RetType)
}

func typeFromRef(rf *ast.TypeRef) Type {
	if rf == nil {
		return T(bytecode.TypeVoid)
	}
	if rf.Name == "array" {
		if rf.Elem == nil {
			return Arr(T(bytecode.TypeInvalid))
		}
		return Arr(typeFromRef(rf.Elem))
	}
	switch rf.Name {
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

func (r *Resolver) errorf(pos token.Position, format string, args ...any) {
	r.errs = append(r.errs, fmt.Errorf("%d:%d: %s", pos.Line, pos.Col, fmt.Sprintf(format, args...)))
}
