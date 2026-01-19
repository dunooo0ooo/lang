package parser

import (
	"testing"

	"github.com/dunooo0ooo/lang/internal/lexer"
)

func TestParseProgramSmoke(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "factorial recursion",
			src: `
fn fact(n: int) -> int {
    if n <= 1 { return 1; }
    return n * fact(n - 1);
}
`,
		},
		{
			name: "if expr",
			src: `
let x: int = if 1 < 2 { 10; } else { 20; };
`,
		},
		{
			name: "while loop and assignment",
			src: `
fn f() -> void {
    let x: int = 0;
    while x < 10 {
        x = x + 1;
    }
    return;
}
`,
		},
		{
			name: "for loop with let init",
			src: `
fn f() -> void {
    let sum: int = 0;
    for let i: int = 0; i < 10; i = i + 1 {
        sum = sum + i;
    }
    return;
}
`,
		},
		{
			name: "float literal and arithmetic",
			src: `
fn f() -> float {
    let x: float = 1.25;
    let y: float = x * 2.0 + 3.5;
    return y;
}
`,
		},
		{
			name: "string and char and null literals",
			src: `
fn f() -> void {
    let s: string = "hello";
    let c: char = 'a';
    let n = null;
    return;
}
`,
		},
		{
			name: "array literal and indexing",
			src: `
fn f() -> int {
    let a: []int = [1, 2, 3];
    let x: int = a[0];
    let y: int = a[1 + 1];
    return x + y;
}
`,
		},
		{
			name: "nested calls and indexing chain",
			src: `
fn id(x: int) -> int { return x; }
fn f() -> int {
    let a: []int = [10, 20, 30];
    return id(a[0]) + id(a[1]);
}
`,
		},
		{
			name: "unary ops and precedence",
			src: `
fn f() -> bool {
    let x: int = 1 + 2 * 3;
    let ok: bool = !(x == 7) || (x != 7) && true;
    return ok;
}
`,
		},
		{
			name: "comments",
			src: `
fn f() -> int {
    // line comment
    let x: int = 1; /* block comment */
    return x;
}
`,
		},
		{
			name: "block tail expr",
			src: `
fn f() -> int {
    let x: int = { let y: int = 2; y + 3 };
    return x;
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.src)
			p := New(l)
			_ = p.ParseProgram()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors:\n%v", p.Errors())
			}
		})
	}
}
