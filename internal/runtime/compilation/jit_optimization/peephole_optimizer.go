package jit_optimization

import "github.com/dunooo0ooo/lang/internal/bytecode"

func OptimizePeephole(fn *bytecode.FunctionInfo) {
	chunk := &fn.Chunk
	originalCode := chunk.Code

	var patches []codePatch

	for instructionPointer := 0; instructionPointer < len(originalCode); {
		op := bytecode.OpCode(originalCode[instructionPointer])
		size := GetInstructionSize(op)

		if size <= 0 || instructionPointer+size > len(originalCode) {
			break
		}

		if found, optimized, patternSize := DetectSwapPattern(originalCode, instructionPointer); found {
			patches = append(patches, codePatch{
				startAddress: instructionPointer,
				endAddress:   instructionPointer + patternSize,
				newCode:      optimized,
			})
			instructionPointer += patternSize
			continue
		}

		instructionPointer += size
	}

	if len(patches) > 0 {
		chunk.Code = rewriteBytecode(originalCode, patches)
	}
}
