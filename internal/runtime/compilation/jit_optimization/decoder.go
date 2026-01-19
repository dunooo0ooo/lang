package jit_optimization

import "github.com/dunooo0ooo/lang/internal/bytecode"

func Decode(code []byte, ip int) (Instruction, bool) {
	if ip >= len(code) {
		return Instruction{}, false
	}

	opCode := bytecode.OpCode(code[ip])
	switch opCode {
	case bytecode.OpConst, bytecode.OpJump, bytecode.OpJumpIfFalse, bytecode.OpCall:
		if ip+2 >= len(code) {
			return Instruction{}, false
		}
		argument := int(uint16(code[ip+1])<<8 | uint16(code[ip+2]))
		return Instruction{OpCode: opCode, Argument: argument, Size: 3}, true

	case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
		if ip+1 >= len(code) {
			return Instruction{}, false
		}
		return Instruction{OpCode: opCode, Argument: int(code[ip+1]), Size: 2}, true

	default:
		return Instruction{OpCode: opCode, Argument: 0, Size: 1}, true
	}
}

func GetInstructionSize(op bytecode.OpCode) int {
	switch op {
	case bytecode.OpConst, bytecode.OpJump, bytecode.OpJumpIfFalse, bytecode.OpCall:
		return 3
	case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
		return 2
	default:
		return 1
	}
}
