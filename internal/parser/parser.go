package parser

import (
	"fmt"
	"strconv"

	"github.com/dunooo0ooo/lang/internal/ast"
	"github.com/dunooo0ooo/lang/internal/lexer"
	"github.com/dunooo0ooo/lang/internal/token"
)

type Parser struct {
	l *lexer.Lexer

	cur  token.Token
	peek token.Token

	errs []error
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.cur = l.NextToken()
	p.peek = l.NextToken()
	return p
}

func (p *Parser) Errors() []error { return p.errs }

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}
	for p.cur.Type != token.EOF {
		it := p.parseItem()
		if it != nil {
			prog.Items = append(prog.Items, it)
			continue
		}
		p.advance()
	}
	return prog
}

func (p *Parser) parseItem() ast.Item {
	if p.cur.Type == token.FN {
		return p.parseFnDecl()
	}
	s := p.parseStmt()
	if s == nil {
		return nil
	}
	return &ast.StmtItem{S: s}
}

func (p *Parser) parseFnDecl() *ast.FnDecl {
	fnPos := p.cur.Pos
	p.expect(token.FN)

	nameTok := p.expect(token.IDENT)
	p.expect(token.LPAREN)

	var params []ast.Param
	if p.cur.Type != token.RPAREN {
		for {
			id := p.expect(token.IDENT)
			p.expect(token.COLON)
			ty := p.parseTypeRef()
			params = append(params, ast.Param{Name: id.Lit, Type: *ty, Pos: id.Pos})

			if p.cur.Type != token.COMMA {
				break
			}
			p.advance()
		}
	}
	p.expect(token.RPAREN)

	var ret *ast.TypeRef
	if p.cur.Type == token.ARROW {
		p.advance()
		ret = p.parseTypeRef()
	}

	body := p.parseBlockStmt()
	if body == nil {
		p.errorf(fnPos, "expected function body")
		return nil
	}

	return &ast.FnDecl{
		FnPos:   fnPos,
		Name:    nameTok.Lit,
		Params:  params,
		RetType: ret,
		Body:    body,
	}
}

func (p *Parser) parseTypeRef() *ast.TypeRef {
	tok := p.cur

	if tok.Type == token.LBRACKET && p.peek.Type == token.RBRACKET {
		lpos := tok.Pos
		p.advance()
		p.advance()
		elem := p.parseTypeRef()
		return &ast.TypeRef{Name: "array", Elem: elem, Pos: lpos}
	}

	switch tok.Type {
	case token.INT_T, token.BOOL_T, token.FLOAT_T, token.STRING_T, token.CHAR_T, token.VOID_T:
		p.advance()
		return &ast.TypeRef{Name: tok.Lit, Pos: tok.Pos}
	default:
		p.errorf(tok.Pos, "expected type, got %v", tok.Type)
		p.advance()
		return &ast.TypeRef{Name: "<?>", Pos: tok.Pos}
	}
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.cur.Type {
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.LET:
		return p.parseLetStmt(true)
	case token.RETURN:
		return p.parseReturnStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.FOR:
		return p.parseForStmt()
	default:
		return p.parseExprOrAssignStmt()
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	lb := p.cur.Pos
	p.expect(token.LBRACE)

	var stmts []ast.Stmt
	var tail ast.Expr

	for p.cur.Type != token.RBRACE && p.cur.Type != token.EOF {
		// 1) сначала обрабатываем "явные statement" по ключевым словам
		switch p.cur.Type {
		case token.LBRACE, token.LET, token.RETURN, token.IF, token.WHILE, token.FOR:
			s := p.parseStmt()
			if s != nil {
				stmts = append(stmts, s)
			} else {
				p.advance()
			}
			continue
		}

		// 2) assignment-statement (IDENT '=' ...)
		if p.cur.Type == token.IDENT && p.peek.Type == token.ASSIGN {
			s := p.parseExprOrAssignStmt()
			if s != nil {
				stmts = append(stmts, s)
			} else {
				p.advance()
			}
			continue
		}

		// 3) иначе это expression: либо ExprStmt (с ';'), либо tail expr (без ';' перед '}')
		if p.startsExpr(p.cur.Type) {
			exprPos := p.cur.Pos
			x := p.parseExpr(precLowest)

			if p.cur.Type == token.SEMICOLON {
				p.advance()
				stmts = append(stmts, &ast.ExprStmt{ExprPos: exprPos, X: x})
				continue
			}

			if p.cur.Type == token.RBRACE {
				tail = x
				break
			}

			p.errorf(p.cur.Pos, "expected ';' or '}', got %v", p.cur.Type)
			if p.cur.Type != token.EOF {
				p.advance()
			}
			continue
		}

		p.errorf(p.cur.Pos, "unexpected token in block: %v", p.cur.Type)
		p.advance()
	}

	p.expect(token.RBRACE)
	return &ast.BlockStmt{Lbrace: lb, Stmts: stmts, Tail: tail}
}

func (p *Parser) startsExpr(t token.Type) bool {
	switch t {
	case token.INT, token.FLOAT, token.STRING, token.CHAR, token.NULL,
		token.TRUE, token.FALSE,
		token.IDENT,
		token.MINUS, token.BANG,
		token.LPAREN,
		token.LBRACE,
		token.IF,
		token.LBRACKET:
		return true
	default:
		return false
	}
}

func (p *Parser) parseLetStmt(withSemi bool) *ast.LetStmt {
	letPos := p.cur.Pos
	p.expect(token.LET)
	nameTok := p.expect(token.IDENT)

	var ty *ast.TypeRef
	if p.cur.Type == token.COLON {
		p.advance()
		ty = p.parseTypeRef()
	}

	var init ast.Expr
	if p.cur.Type == token.ASSIGN {
		p.advance()
		init = p.parseExpr(precLowest)
	}

	if withSemi {
		p.expect(token.SEMICOLON)
	}

	return &ast.LetStmt{LetPos: letPos, Name: nameTok.Lit, Type: ty, Init: init}
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	pos := p.cur.Pos
	p.expect(token.RETURN)

	var val ast.Expr
	if p.cur.Type != token.SEMICOLON {
		val = p.parseExpr(precLowest)
	}

	p.expect(token.SEMICOLON)
	return &ast.ReturnStmt{RetPos: pos, Value: val}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	pos := p.cur.Pos
	p.expect(token.IF)

	cond := p.parseExpr(precLowest)
	thenBlk := p.parseBlockStmt()

	var els ast.Stmt
	if p.cur.Type == token.ELSE {
		p.advance()
		if p.cur.Type == token.IF {
			els = p.parseIfStmt()
		} else {
			els = p.parseBlockStmt()
		}
	}

	return &ast.IfStmt{IfPos: pos, Cond: cond, Then: thenBlk, Else: els}
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	pos := p.cur.Pos
	p.expect(token.WHILE)

	cond := p.parseExpr(precLowest)
	body := p.parseBlockStmt()
	return &ast.WhileStmt{WhilePos: pos, Cond: cond, Body: body}
}

func (p *Parser) parseForStmt() *ast.ForStmt {
	pos := p.cur.Pos
	p.expect(token.FOR)

	var init ast.Stmt
	if p.cur.Type != token.SEMICOLON {
		if p.cur.Type == token.LET {
			init = p.parseLetStmt(false)
		} else {
			init = p.parseAssignOrExprStmt(false)
		}
	}
	p.expect(token.SEMICOLON)

	var cond ast.Expr
	if p.cur.Type != token.SEMICOLON {
		cond = p.parseExpr(precLowest)
	}
	p.expect(token.SEMICOLON)

	var post ast.Stmt
	if p.cur.Type != token.LBRACE {
		post = p.parseAssignOrExprStmt(false)
	}

	body := p.parseBlockStmt()
	return &ast.ForStmt{ForPos: pos, Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseExprOrAssignStmt() ast.Stmt {
	if p.cur.Type == token.IDENT && p.peek.Type == token.ASSIGN {
		nameTok := p.cur
		p.advance()
		p.advance()
		val := p.parseExpr(precLowest)
		p.expect(token.SEMICOLON)
		return &ast.AssignStmt{NamePos: nameTok.Pos, Name: nameTok.Lit, Value: val}
	}

	exprPos := p.cur.Pos
	x := p.parseExpr(precLowest)
	p.expect(token.SEMICOLON)
	return &ast.ExprStmt{ExprPos: exprPos, X: x}
}

func (p *Parser) parseAssignOrExprStmt(withSemi bool) ast.Stmt {
	if p.cur.Type == token.IDENT && p.peek.Type == token.ASSIGN {
		nameTok := p.cur
		p.advance()
		p.advance()
		val := p.parseExpr(precLowest)
		if withSemi {
			p.expect(token.SEMICOLON)
		}
		return &ast.AssignStmt{NamePos: nameTok.Pos, Name: nameTok.Lit, Value: val}
	}

	exprPos := p.cur.Pos
	x := p.parseExpr(precLowest)
	if withSemi {
		p.expect(token.SEMICOLON)
	}
	return &ast.ExprStmt{ExprPos: exprPos, X: x}
}

func (p *Parser) parseExpr(min prec) ast.Expr {
	var left ast.Expr

	switch p.cur.Type {
	case token.INT:
		left = p.parseIntLit()
	case token.FLOAT:
		left = p.parseFloatLit()
	case token.STRING:
		left = &ast.StringLit{Pos0: p.cur.Pos, Value: p.cur.Lit}
		p.advance()
	case token.CHAR:
		left = &ast.CharLit{Pos0: p.cur.Pos, Raw: p.cur.Lit}
		p.advance()
	case token.NULL:
		left = &ast.NullLit{Pos0: p.cur.Pos}
		p.advance()
	case token.TRUE, token.FALSE:
		left = p.parseBoolLit()
	case token.IDENT:
		left = &ast.VarRef{NamePos: p.cur.Pos, Name: p.cur.Lit}
		p.advance()
	case token.MINUS, token.BANG:
		opTok := p.cur
		p.advance()
		x := p.parseExpr(precUnary)
		left = &ast.UnaryExpr{OpPos: opTok.Pos, Op: opTok.Type, X: x}
	case token.LPAREN:
		p.advance()
		left = p.parseExpr(precLowest)
		p.expect(token.RPAREN)
	case token.LBRACE:
		blk := p.parseBlockStmt()
		left = &ast.BlockExpr{Block: blk}
	case token.IF:
		left = p.parseIfExpr()
	case token.LBRACKET:
		left = p.parseArrayLit()
	default:
		p.errorf(p.cur.Pos, "unexpected token in expression: %v", p.cur.Type)
		p.advance()
		left = &ast.VarRef{NamePos: p.cur.Pos, Name: "<error>"}
	}

	for p.cur.Type != token.EOF {
		opPrec, ok := precedences[p.cur.Type]
		if !ok || opPrec < min {
			break
		}

		switch p.cur.Type {
		case token.LPAREN:

			left = p.parseCall(left)
		case token.LBRACKET:

			left = p.parseIndex(left)
		default:

			opTok := p.cur
			p.advance()
			right := p.parseExpr(opPrec + 1)
			left = &ast.BinaryExpr{OpPos: opTok.Pos, Op: opTok.Type, L: left, R: right}
		}
	}

	return left
}

func (p *Parser) parseIntLit() ast.Expr {
	tok := p.cur
	p.advance()
	v, err := strconv.ParseInt(tok.Lit, 10, 64)
	if err != nil {
		p.errorf(tok.Pos, "bad int literal: %q", tok.Lit)
		v = 0
	}
	return &ast.IntLit{IntPos: tok.Pos, Value: v, Raw: tok.Lit}
}

func (p *Parser) parseFloatLit() ast.Expr {
	tok := p.cur
	p.advance()
	v, err := strconv.ParseFloat(tok.Lit, 64)
	if err != nil {
		p.errorf(tok.Pos, "bad float literal: %q", tok.Lit)
		v = 0
	}
	return &ast.FloatLit{Pos0: tok.Pos, Value: v, Raw: tok.Lit}
}

func (p *Parser) parseBoolLit() ast.Expr {
	tok := p.cur
	p.advance()
	return &ast.BoolLit{BoolPos: tok.Pos, Value: tok.Type == token.TRUE}
}

func (p *Parser) parseArrayLit() ast.Expr {
	lb := p.cur.Pos
	p.expect(token.LBRACKET)

	var elems []ast.Expr
	if p.cur.Type != token.RBRACKET {
		for {
			elems = append(elems, p.parseExpr(precLowest))
			if p.cur.Type != token.COMMA {
				break
			}
			p.advance()
		}
	}

	p.expect(token.RBRACKET)
	return &ast.ArrayLit{Lbrack: lb, Elems: elems}
}

func (p *Parser) parseIndex(x ast.Expr) ast.Expr {
	lb := p.cur.Pos
	p.expect(token.LBRACKET)
	idx := p.parseExpr(precLowest)
	p.expect(token.RBRACKET)
	return &ast.IndexExpr{Lbrack: lb, X: x, Index: idx}
}

func (p *Parser) parseCall(callee ast.Expr) ast.Expr {
	lp := p.cur.Pos
	p.expect(token.LPAREN)

	var args []ast.Expr
	if p.cur.Type != token.RPAREN {
		for {
			args = append(args, p.parseExpr(precLowest))
			if p.cur.Type != token.COMMA {
				break
			}
			p.advance()
		}
	}

	p.expect(token.RPAREN)
	return &ast.CallExpr{Lparen: lp, Callee: callee, Args: args}
}

func (p *Parser) parseIfExpr() ast.Expr {
	pos := p.cur.Pos
	p.expect(token.IF)

	cond := p.parseExpr(precLowest)
	thenBlk := p.parseBlockStmt()

	p.expect(token.ELSE)

	var els ast.Expr
	if p.cur.Type == token.IF {
		els = p.parseIfExpr()
	} else if p.cur.Type == token.LBRACE {
		els = &ast.BlockExpr{Block: p.parseBlockStmt()}
	} else {
		p.errorf(p.cur.Pos, "expected else block or else-if")
		els = &ast.VarRef{NamePos: p.cur.Pos, Name: "<error>"}
	}

	return &ast.IfExpr{IfPos: pos, Cond: cond, Then: thenBlk, Else: els}
}

func (p *Parser) advance() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) expect(t token.Type) token.Token {
	if p.cur.Type != t {
		p.errorf(p.cur.Pos, "expected %v, got %v (%q)", t, p.cur.Type, p.cur.Lit)
		got := p.cur
		p.advance()
		return got
	}
	got := p.cur
	p.advance()
	return got
}

func (p *Parser) errorf(pos token.Position, format string, args ...any) {
	p.errs = append(p.errs, fmt.Errorf("%d:%d: %s", pos.Line, pos.Col, fmt.Sprintf(format, args...)))
}
