package sema

import (
	"testing"

	"github.com/dunooo0ooo/lang/internal/lexer"
	"github.com/dunooo0ooo/lang/internal/parser"
)

func TestSemaSmoke_AllTypes(t *testing.T) {
	src := `
fn fact(n: int) -> int {
    if n <= 1 { return 1; }
    return n * fact(n - 1);
}

fn f() -> void {
    let x: float = 1.25;
    let s: string = "hi";
    let c: char = 'a';
    let a: []int = [1,2,3];
    let y: int = a[0];
    let z = null;
    let xi: int = if 1 < 2 { 10 } else { 20 };
    return;
}
`
	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	c.Check(prog)
	if len(c.Errors()) != 0 {
		t.Fatalf("sema errors: %v", c.Errors())
	}
}
