package compilation

import (
	"fmt"

	"github.com/dunooo0ooo/lang/internal/ast"
	"github.com/dunooo0ooo/lang/internal/bytecode"
	"github.com/dunooo0ooo/lang/internal/token"
)

func (c *Compiler) CompileProgram(p *ast.Program) (*bytecode.Module, error) {
	for _, it := range p.Items {
		fn, ok := it.(*ast.FnDecl)
		if !ok {
			continue
		}
		if _, exists := c.mod.Functions[fn.Name]; exists {
			return nil, fmt.Errorf("duplicate function: %s", fn.Name)
		}

		bfn := bytecode.CreateFunction(fn.Name, len(fn.Params))

		for _, par := range fn.Params {
			bfn.AddParameter(mapTypeRef(&par.Type))
		}

		ret := bytecode.TypeVoid
		if fn.RetType != nil {
			ret = mapTypeRef(fn.RetType)
		}
		bfn.SetReturnType(ret)

		c.mod.Functions[bfn.Name] = bfn
	}

	for _, it := range p.Items {
		fn, ok := it.(*ast.FnDecl)
		if !ok {
			continue
		}
		if err := c.compileFunction(fn); err != nil {
			return nil, err
		}
	}

	return c.mod, nil
}

func mapTypeRef(t *ast.TypeRef) bytecode.TypeKind {
	if t == nil {
		return bytecode.TypeVoid
	}
	if t.Name == "array" {
		return bytecode.TypeArray
	}
	switch t.Name {
	case "int":
		return bytecode.TypeInt
	case "float":
		return bytecode.TypeFloat
	case "bool":
		return bytecode.TypeBool
	case "string":
		return bytecode.TypeString
	case "char":
		return bytecode.TypeChar
	case "void":
		return bytecode.TypeVoid
	case "null":
		return bytecode.TypeNull
	default:
		return bytecode.TypeInvalid
	}
}

func (c *Compiler) compileFunction(fn *ast.FnDecl) error {
	bfn, ok := c.mod.Functions[fn.Name]
	if !ok {
		return fmt.Errorf("function %s not registered", fn.Name)
	}

	c.fn = bfn
	c.locals = nil

	bfn.Chunk = bytecode.Chunk{}
	bfn.NumLocals = 0

	for _, p := range fn.Params {
		c.addLocal(p.Name, mapTypeRef(&p.Type))
	}

	c.compileBlock(fn.Body, false)

	c.emitNull()
	c.chunk().Write(bytecode.OpReturn)

	return nil
}

func (c *Compiler) compileBlock(b *ast.BlockStmt, asExpr bool) {
	for _, st := range b.Stmts {
		c.compileStmt(st)
	}

	if b.Tail != nil {
		c.compileExpr(b.Tail)
		if !asExpr {
			c.chunk().Write(bytecode.OpPop)
		}
		return
	}

	if asExpr {
		c.emitNull()
	}
}

func (c *Compiler) compileStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.BlockStmt:
		c.compileBlock(st, false)

	case *ast.LetStmt:
		c.compileLet(st)

	case *ast.AssignStmt:
		c.compileAssign(st)

	case *ast.ExprStmt:
		c.compileExpr(st.X)
		c.chunk().Write(bytecode.OpPop)

	case *ast.ReturnStmt:
		c.compileReturn(st)

	case *ast.IfStmt:
		c.compileIfStmt(st)

	case *ast.WhileStmt:
		c.compileWhile(st)

	case *ast.ForStmt:
		c.compileFor(st)

	default:
		panic(fmt.Sprintf("unknown stmt %T", st))
	}
}

func (c *Compiler) compileLet(s *ast.LetStmt) {
	ch := c.chunk()

	if s.Init != nil {
		c.compileExpr(s.Init)
	} else {
		c.emitNull()
	}

	typ := bytecode.TypeInvalid
	if s.Type != nil {
		typ = mapTypeRef(s.Type)
	}

	slot := c.addLocal(s.Name, typ)

	ch.Write(bytecode.OpStoreLocal)
	_ = ch.WriteByte(byte(slot))
}

func (c *Compiler) compileAssign(s *ast.AssignStmt) {
	ch := c.chunk()

	c.compileExpr(s.Value)

	slot, ok := c.resolveLocal(s.Name)
	if !ok {
		panic("unknown variable " + s.Name)
	}

	ch.Write(bytecode.OpStoreLocal)
	_ = ch.WriteByte(byte(slot))
}

func (c *Compiler) compileReturn(s *ast.ReturnStmt) {
	ch := c.chunk()
	if s.Value != nil {
		c.compileExpr(s.Value)
	} else {
		c.emitNull()
	}
	ch.Write(bytecode.OpReturn)
}

func (c *Compiler) compileIfStmt(s *ast.IfStmt) {
	ch := c.chunk()

	c.compileExpr(s.Cond)

	ch.Write(bytecode.OpJumpIfFalse)
	jumpToElse := len(ch.Code)
	ch.WriteUint16(0)

	ch.Write(bytecode.OpPop)
	c.compileBlock(s.Then, false)

	ch.Write(bytecode.OpJump)
	jumpAfterElse := len(ch.Code)
	ch.WriteUint16(0)

	elsePos := len(ch.Code)
	_ = ch.PatchUint16(jumpToElse, uint16(elsePos))

	ch.Write(bytecode.OpPop)
	if s.Else != nil {
		c.compileStmt(s.Else)
	}

	endPos := len(ch.Code)
	_ = ch.PatchUint16(jumpAfterElse, uint16(endPos))
}

func (c *Compiler) compileWhile(s *ast.WhileStmt) {
	ch := c.chunk()

	loopStart := len(ch.Code)
	c.beginLoop()

	c.compileExpr(s.Cond)

	ch.Write(bytecode.OpJumpIfFalse)
	exitJump := len(ch.Code)
	ch.WriteUint16(0)

	ch.Write(bytecode.OpPop)
	c.compileBlock(s.Body, false)

	ch.Write(bytecode.OpJump)
	ch.WriteUint16(uint16(loopStart))

	exitLabel := len(ch.Code)
	_ = ch.PatchUint16(exitJump, uint16(exitLabel))
	ch.Write(bytecode.OpPop)

	afterLoop := len(ch.Code)
	c.endLoop(loopStart, afterLoop)
}

func (c *Compiler) compileFor(s *ast.ForStmt) {
	ch := c.chunk()

	if s.Init != nil {
		c.compileStmt(s.Init)
	}

	loopStart := len(ch.Code)
	c.beginLoop()

	hasCond := s.Cond != nil
	if hasCond {
		c.compileExpr(s.Cond)

		ch.Write(bytecode.OpJumpIfFalse)
		exitJump := len(ch.Code)
		ch.WriteUint16(0)

		ch.Write(bytecode.OpPop)

		c.compileBlock(s.Body, false)

		continueTarget := loopStart
		if s.Post != nil {
			continueTarget = len(ch.Code)
			c.compileStmt(s.Post)
		}

		ch.Write(bytecode.OpJump)
		ch.WriteUint16(uint16(loopStart))

		exitLabel := len(ch.Code)
		_ = ch.PatchUint16(exitJump, uint16(exitLabel))
		ch.Write(bytecode.OpPop)

		afterLoop := len(ch.Code)
		c.endLoop(continueTarget, afterLoop)
		return
	}

	c.compileBlock(s.Body, false)

	continueTarget := loopStart
	if s.Post != nil {
		continueTarget = len(ch.Code)
		c.compileStmt(s.Post)
	}

	ch.Write(bytecode.OpJump)
	ch.WriteUint16(uint16(loopStart))

	afterLoop := len(ch.Code)
	c.endLoop(continueTarget, afterLoop)
}

func (c *Compiler) compileExpr(e ast.Expr) {
	switch ex := e.(type) {
	case *ast.VarRef:
		c.compileIdent(ex)

	case *ast.IntLit:
		c.emitInt(ex.Value)

	case *ast.FloatLit:
		c.emitFloat(ex.Value)

	case *ast.BoolLit:
		c.emitBool(ex.Value)

	case *ast.StringLit:
		c.emitString(ex.Value)

	case *ast.CharLit:
		var b byte
		if len(ex.Raw) > 0 {
			b = ex.Raw[0]
		}
		c.emitChar(b)

	case *ast.NullLit:
		c.emitNull()

	case *ast.UnaryExpr:
		c.compileExpr(ex.X)
		switch ex.Op {
		case token.MINUS:
			c.chunk().Write(bytecode.OpNeg)
		case token.BANG:
			c.chunk().Write(bytecode.OpNot)
		default:
			panic("unknown unary op")
		}

	case *ast.BinaryExpr:
		c.compileBinary(ex)

	case *ast.CallExpr:
		c.compileCall(ex)

	case *ast.BlockExpr:
		c.compileBlock(ex.Block, true)

	case *ast.IfExpr:
		c.compileIfExpr(ex)

	case *ast.ArrayLit:
		c.compileArrayLit(ex)

	case *ast.IndexExpr:
		c.compileExpr(ex.X)
		c.compileExpr(ex.Index)
		c.chunk().Write(bytecode.OpArrayGet)

	default:
		panic(fmt.Sprintf("unknown expr %T", ex))
	}
}

func (c *Compiler) compileBinary(e *ast.BinaryExpr) {
	ch := c.chunk()

	switch e.Op {
	case token.AND:
		c.compileExpr(e.L)

		ch.Write(bytecode.OpJumpIfFalse)
		jumpToEnd := len(ch.Code)
		ch.WriteUint16(0)

		ch.Write(bytecode.OpPop)
		c.compileExpr(e.R)

		end := len(ch.Code)
		_ = ch.PatchUint16(jumpToEnd, uint16(end))
		return

	case token.OR:
		c.compileExpr(e.L)

		ch.Write(bytecode.OpJumpIfFalse)
		jumpToRight := len(ch.Code)
		ch.WriteUint16(0)

		ch.Write(bytecode.OpJump)
		jumpAfterTrue := len(ch.Code)
		ch.WriteUint16(0)

		rightPos := len(ch.Code)
		_ = ch.PatchUint16(jumpToRight, uint16(rightPos))

		ch.Write(bytecode.OpPop)
		c.compileExpr(e.R)

		end := len(ch.Code)
		_ = ch.PatchUint16(jumpAfterTrue, uint16(end))
		return
	}

	c.compileExpr(e.L)
	c.compileExpr(e.R)

	switch e.Op {
	case token.PLUS:
		ch.Write(bytecode.OpAdd)
	case token.MINUS:
		ch.Write(bytecode.OpSub)
	case token.STAR:
		ch.Write(bytecode.OpMul)
	case token.SLASH:
		ch.Write(bytecode.OpDiv)
	case token.PERCENT:
		ch.Write(bytecode.OpMod)

	case token.EQ:
		ch.Write(bytecode.OpEq)
	case token.NEQ:
		ch.Write(bytecode.OpNe)
	case token.LT:
		ch.Write(bytecode.OpLt)
	case token.LTE:
		ch.Write(bytecode.OpLe)
	case token.GT:
		ch.Write(bytecode.OpGt)
	case token.GTE:
		ch.Write(bytecode.OpGe)

	default:
		panic("unknown binary op: " + e.Op.String())
	}
}

func (c *Compiler) compileIfExpr(e *ast.IfExpr) {
	ch := c.chunk()

	c.compileExpr(e.Cond)

	ch.Write(bytecode.OpJumpIfFalse)
	jumpToElse := len(ch.Code)
	ch.WriteUint16(0)

	ch.Write(bytecode.OpPop)
	c.compileBlock(e.Then, true)

	ch.Write(bytecode.OpJump)
	jumpAfterElse := len(ch.Code)
	ch.WriteUint16(0)

	elsePos := len(ch.Code)
	_ = ch.PatchUint16(jumpToElse, uint16(elsePos))

	ch.Write(bytecode.OpPop)
	c.compileExpr(e.Else)

	endPos := len(ch.Code)
	_ = ch.PatchUint16(jumpAfterElse, uint16(endPos))
}

func (c *Compiler) compileCall(e *ast.CallExpr) {
	ch := c.chunk()

	id, ok := e.Callee.(*ast.VarRef)
	if !ok {
		panic("call of non-identifier is not supported")
	}

	for _, arg := range e.Args {
		c.compileExpr(arg)
	}

	name := id.Name

	if name == "print" {
		if len(e.Args) != 1 {
			panic("print expects exactly 1 argument")
		}
		ch.Write(bytecode.OpPrint)
		c.emitNull()
		return
	}

	if _, ok := c.mod.Functions[name]; !ok {
		panic("unknown function: " + name)
	}

	ch.Write(bytecode.OpCall)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValString, S: name})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) compileIdent(e *ast.VarRef) {
	ch := c.chunk()
	if slot, ok := c.resolveLocal(e.Name); ok {
		ch.Write(bytecode.OpLoadLocal)
		_ = ch.WriteByte(byte(slot))
		return
	}
	panic("unknown variable: " + e.Name)
}

func (c *Compiler) compileArrayLit(a *ast.ArrayLit) {
	ch := c.chunk()

	c.emitInt(int64(len(a.Elems)))
	ch.Write(bytecode.OpArrayNew)

	tmpSlot := c.addLocal("$tmp_arr", bytecode.TypeArray)
	ch.Write(bytecode.OpStoreLocal)
	_ = ch.WriteByte(byte(tmpSlot))

	for i, el := range a.Elems {
		ch.Write(bytecode.OpLoadLocal)
		_ = ch.WriteByte(byte(tmpSlot))

		c.emitInt(int64(i))
		c.compileExpr(el)
		ch.Write(bytecode.OpArraySet)
	}

	ch.Write(bytecode.OpLoadLocal)
	_ = ch.WriteByte(byte(tmpSlot))
}

func (c *Compiler) emitInt(v int64) {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValInt, I: v})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) emitFloat(v float64) {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValFloat, F: v})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) emitBool(v bool) {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValBool, B: v})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) emitString(s string) {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValString, S: s})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) emitChar(b byte) {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValChar, C: b})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) emitNull() {
	ch := c.chunk()
	ch.Write(bytecode.OpConst)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
	ch.WriteUint16(uint16(idx))
}

func (c *Compiler) beginLoop() {
	c.breakStack = append(c.breakStack, nil)
	c.continueStack = append(c.continueStack, nil)
}

func (c *Compiler) endLoop(continueTarget, breakTarget int) {
	ch := c.chunk()

	bi := len(c.breakStack) - 1
	for _, pos := range c.breakStack[bi] {
		_ = ch.PatchUint16(pos, uint16(breakTarget))
	}
	c.breakStack = c.breakStack[:bi]

	ci := len(c.continueStack) - 1
	for _, pos := range c.continueStack[ci] {
		_ = ch.PatchUint16(pos, uint16(continueTarget))
	}
	c.continueStack = c.continueStack[:ci]
}
