package jit_optimization

import "github.com/dunooo0ooo/lang/internal/bytecode"

type Instruction struct {
	OpCode   bytecode.OpCode
	Argument int
	Size     int
}