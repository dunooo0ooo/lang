package parser

import "github.com/dunooo0ooo/lang/internal/token"

type prec int

const (
	precLowest prec = iota
	precOr
	precAnd
	precEq
	precCmp
	precAdd
	precMul
	precUnary
	precCall
)

var precedences = map[token.Type]prec{
	token.OR:  precOr,
	token.AND: precAnd,

	token.EQ:  precEq,
	token.NEQ: precEq,

	token.LT:  precCmp,
	token.LTE: precCmp,
	token.GT:  precCmp,
	token.GTE: precCmp,

	token.PLUS:  precAdd,
	token.MINUS: precAdd,

	token.STAR:    precMul,
	token.SLASH:   precMul,
	token.PERCENT: precMul,

	token.LPAREN:   precCall,
	token.LBRACKET: precCall,
}
