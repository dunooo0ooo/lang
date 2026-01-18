package lexer

import (
	"github.com/dunooo0ooo/lang/internal/token"
	"testing"
)

func TestLexerBasic(t *testing.T) {
	src := `
fn fact(n: int) -> int {
    if n <= 1 { return 1; }
    return n * fact(n - 1);
}
`
	l := New(src)

	for {
		tok := l.NextToken()
		if tok.Type == token.ILLEGAL {
			t.Fatalf("illegal token at %v: %q", tok.Pos, tok.Lit)
		}
		if tok.Type == token.EOF {
			break
		}
	}
}
