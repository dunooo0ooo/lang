package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dunooo0ooo/lang/internal/bytecode"
	"github.com/dunooo0ooo/lang/internal/lexer"
	"github.com/dunooo0ooo/lang/internal/parser"
	"github.com/dunooo0ooo/lang/internal/runtime"
	"github.com/dunooo0ooo/lang/internal/runtime/compilation"
	"github.com/dunooo0ooo/lang/internal/sema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: langrun <file.lang> [--jit]")
		os.Exit(1)
	}

	path := os.Args[1]
	enableJit := len(os.Args) > 2 && os.Args[2] == "--jit"

	src, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	p := parser.New(lexer.New(string(src)))
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, e := range p.Errors() {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	s := sema.New()
	s.Check(prog)
	if len(s.Errors()) != 0 {
		for _, e := range s.Errors() {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	comp := compilation.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		panic(err)
	}

	vm := runtime.NewVM(mod, enableJit)

	start := time.Now()
	ret, err := vm.Call("main", nil)
	if err != nil {
		panic(err)
	}
	elapsed := time.Since(start)

	printResult(ret)
	fmt.Println("time:", elapsed)
}

func printResult(v bytecode.Value) {
	switch v.Kind {
	case bytecode.ValInt:
		fmt.Println("result:", v.I)
	case bytecode.ValFloat:
		fmt.Println("result:", v.F)
	default:
		fmt.Println("result:", v)
	}
}
