package sema

import (
	"testing"

	"github.com/dunooo0ooo/lang/internal/lexer"
	"github.com/dunooo0ooo/lang/internal/parser"
)

func TestResolverSlots(t *testing.T) {
	src := `
fn f(a: int, b: int) -> int {
    let x: int = a + b;
    x = x + 1;
    return x;
}
`
	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	ch := New()
	ch.Check(prog)
	if len(ch.Errors()) != 0 {
		t.Fatalf("sema errors: %v", ch.Errors())
	}

	r := NewResolver(ch.ExprType)
	r.Resolve(prog)
	if len(r.Errors()) != 0 {
		t.Fatalf("resolver errors: %v", r.Errors())
	}

	if len(r.Result().Fns) != 1 {
		t.Fatalf("expected 1 function, got %d", len(r.Result().Fns))
	}
}
