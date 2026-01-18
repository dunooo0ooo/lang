package ast

import "github.com/dunooo0ooo/lang/internal/token"

type Node interface {
	Pos() token.Position
}

type Program struct {
	Items []Item
}

type Item interface {
	Node
	isItem()
}

type StmtItem struct {
	S Stmt
}

func (i *StmtItem) Pos() token.Position { return i.S.Pos() }
func (i *StmtItem) isItem()             {}

type FnDecl struct {
	FnPos   token.Position
	Name    string
	Params  []Param
	RetType *TypeRef
	Body    *BlockStmt
}

func (d *FnDecl) Pos() token.Position { return d.FnPos }
func (d *FnDecl) isItem()             {}

type Param struct {
	Name string
	Type TypeRef
	Pos  token.Position
}

type TypeRef struct {
	Name string
	Elem *TypeRef
	Pos  token.Position
}

type Stmt interface {
	Node
	isStmt()
}

type BlockStmt struct {
	Lbrace token.Position
	Stmts  []Stmt
}

func (s *BlockStmt) Pos() token.Position { return s.Lbrace }
func (s *BlockStmt) isStmt()             {}

type LetStmt struct {
	LetPos token.Position
	Name   string
	Type   *TypeRef
	Init   Expr
}

func (s *LetStmt) Pos() token.Position { return s.LetPos }
func (s *LetStmt) isStmt()             {}

type AssignStmt struct {
	NamePos token.Position
	Name    string
	Value   Expr
}

func (s *AssignStmt) Pos() token.Position { return s.NamePos }
func (s *AssignStmt) isStmt()             {}

type ReturnStmt struct {
	RetPos token.Position
	Value  Expr
}

func (s *ReturnStmt) Pos() token.Position { return s.RetPos }
func (s *ReturnStmt) isStmt()             {}

type IfStmt struct {
	IfPos token.Position
	Cond  Expr
	Then  *BlockStmt
	Else  Stmt
}

func (s *IfStmt) Pos() token.Position { return s.IfPos }
func (s *IfStmt) isStmt()             {}

type WhileStmt struct {
	WhilePos token.Position
	Cond     Expr
	Body     *BlockStmt
}

func (s *WhileStmt) Pos() token.Position { return s.WhilePos }
func (s *WhileStmt) isStmt()             {}

type ForStmt struct {
	ForPos token.Position
	Init   Stmt
	Cond   Expr
	Post   Stmt
	Body   *BlockStmt
}

func (s *ForStmt) Pos() token.Position { return s.ForPos }
func (s *ForStmt) isStmt()             {}

type ExprStmt struct {
	ExprPos token.Position
	X       Expr
}

func (s *ExprStmt) Pos() token.Position { return s.ExprPos }
func (s *ExprStmt) isStmt()             {}

type Expr interface {
	Node
	isExpr()
}

type IntLit struct {
	IntPos token.Position
	Value  int64
	Raw    string
}

func (e *IntLit) Pos() token.Position { return e.IntPos }
func (e *IntLit) isExpr()             {}

type BoolLit struct {
	BoolPos token.Position
	Value   bool
}

func (e *BoolLit) Pos() token.Position { return e.BoolPos }
func (e *BoolLit) isExpr()             {}

type VarRef struct {
	NamePos token.Position
	Name    string
}

func (e *VarRef) Pos() token.Position { return e.NamePos }
func (e *VarRef) isExpr()             {}

type UnaryExpr struct {
	OpPos token.Position
	Op    token.Type
	X     Expr
}

func (e *UnaryExpr) Pos() token.Position { return e.OpPos }
func (e *UnaryExpr) isExpr()             {}

type BinaryExpr struct {
	OpPos token.Position
	Op    token.Type
	L     Expr
	R     Expr
}

func (e *BinaryExpr) Pos() token.Position { return e.OpPos }
func (e *BinaryExpr) isExpr()             {}

type CallExpr struct {
	Lparen token.Position
	Callee Expr
	Args   []Expr
}

func (e *CallExpr) Pos() token.Position { return e.Lparen }
func (e *CallExpr) isExpr()             {}

type IfExpr struct {
	IfPos token.Position
	Cond  Expr
	Then  *BlockStmt
	Else  Expr
}

func (e *IfExpr) Pos() token.Position { return e.IfPos }
func (e *IfExpr) isExpr()             {}

type BlockExpr struct {
	Block *BlockStmt
}

func (e *BlockExpr) Pos() token.Position { return e.Block.Pos() }
func (e *BlockExpr) isExpr()             {}

type FloatLit struct {
	Pos0  token.Position
	Value float64
	Raw   string
}

func (e *FloatLit) Pos() token.Position { return e.Pos0 }
func (e *FloatLit) isExpr()             {}

type StringLit struct {
	Pos0  token.Position
	Value string
}

func (e *StringLit) Pos() token.Position { return e.Pos0 }
func (e *StringLit) isExpr()             {}

type CharLit struct {
	Pos0 token.Position
	Raw  string
}

func (e *CharLit) Pos() token.Position { return e.Pos0 }
func (e *CharLit) isExpr()             {}

type NullLit struct {
	Pos0 token.Position
}

func (e *NullLit) Pos() token.Position { return e.Pos0 }
func (e *NullLit) isExpr()             {}

type ArrayLit struct {
	Lbrack token.Position
	Elems  []Expr
}

func (e *ArrayLit) Pos() token.Position { return e.Lbrack }
func (e *ArrayLit) isExpr()             {}

type IndexExpr struct {
	Lbrack token.Position
	X      Expr
	Index  Expr
}

func (e *IndexExpr) Pos() token.Position { return e.Lbrack }
func (e *IndexExpr) isExpr()             {}
