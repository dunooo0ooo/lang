package e2e_test

import (
	"testing"

	"github.com/dunooo0ooo/lang/internal/ast"
	"github.com/dunooo0ooo/lang/internal/bytecode"
	"github.com/dunooo0ooo/lang/internal/lexer"
	"github.com/dunooo0ooo/lang/internal/parser"
	"github.com/dunooo0ooo/lang/internal/runtime"
	"github.com/dunooo0ooo/lang/internal/runtime/compilation"
	"github.com/dunooo0ooo/lang/internal/sema"
)

func TestE2E_Factorial(t *testing.T) {
	src := `
fn fact(n: int) -> int {
    if n <= 1 { return 1; }
    return n * fact(n - 1);
}
`
	prog := mustParse(t, src)
	mustSema(t, prog)

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := runtime.NewVM(mod, false)
	ret, err := vm.Call("fact", []bytecode.Value{{Kind: bytecode.ValInt, I: 5}})
	if err != nil {
		t.Fatalf("vm call error: %v", err)
	}
	if ret.Kind != bytecode.ValInt || ret.I != 120 {
		t.Fatalf("unexpected result: %#v, want int 120", ret)
	}
}

func TestE2E_IfExpr(t *testing.T) {
	src := `
fn main() -> int {
    let x: int = if 1 < 2 { 10 } else { 20 };
    return x;
}
`
	prog := mustParse(t, src)
	mustSema(t, prog)

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := runtime.NewVM(mod, false)
	ret, err := vm.Call("main", nil)
	if err != nil {
		t.Fatalf("vm call error: %v", err)
	}
	if ret.Kind != bytecode.ValInt || ret.I != 10 {
		t.Fatalf("unexpected result: %#v, want int 10", ret)
	}
}

func TestE2E_WhileAndAssign(t *testing.T) {
	src := `
fn main() -> int {
    let i: int = 0;
    let sum: int = 0;
    while i < 5 {
        sum = sum + i;
        i = i + 1;
    }
    return sum;
}
`
	prog := mustParse(t, src)
	mustSema(t, prog)

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := runtime.NewVM(mod, false)
	ret, err := vm.Call("main", nil)
	if err != nil {
		t.Fatalf("vm call error: %v", err)
	}
	// 0+1+2+3+4 = 10
	if ret.Kind != bytecode.ValInt || ret.I != 10 {
		t.Fatalf("unexpected result: %#v, want int 10", ret)
	}
}

func TestE2E_ArrayLitAndIndex(t *testing.T) {
	src := `
fn main() -> int {
    let a: []int = [1,2,3];
    return a[1];
}
`
	prog := mustParse(t, src)
	mustSema(t, prog)

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := runtime.NewVM(mod, false)
	ret, err := vm.Call("main", nil)
	if err != nil {
		t.Fatalf("vm call error: %v", err)
	}
	if ret.Kind != bytecode.ValInt || ret.I != 2 {
		t.Fatalf("unexpected result: %#v, want int 2", ret)
	}
}

func TestE2E_GC_ArrayAllocStress(t *testing.T) {
	src := `
fn main() -> int {
    let i: int = 0;
    let sum: int = 0;

    while i < 200 {
        let a: []int = [1,2,3,4,5,6,7,8,9,10];
        sum = sum + a[0];
        i = i + 1;
    }

    return sum;
}
`
	prog := mustParse(t, src)
	mustSema(t, prog)

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := runtime.NewVM(mod, false)

	// Чтобы гарантированно триггерить GC почаще:
	// (тест в package e2e_test, поэтому MaxObjects не виден — оставь как есть)
	ret, err := vm.Call("main", nil)
	if err != nil {
		t.Fatalf("vm call error: %v", err)
	}

	// 200 итераций * a[0]=1 => 200
	if ret.Kind != bytecode.ValInt || ret.I != 200 {
		t.Fatalf("unexpected result: %#v, want int 200", ret)
	}
}

// ---- helpers ----

func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()

	p := parser.New(lexer.New(src))
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors:\n%v", p.Errors())
	}
	return prog
}

func mustSema(t *testing.T, prog *ast.Program) {
	t.Helper()

	c := sema.New()
	c.Check(prog)
	if len(c.Errors()) != 0 {
		t.Fatalf("sema errors:\n%v", c.Errors())
	}
}
