package jit_optimization

import "github.com/dunooo0ooo/lang/internal/bytecode"

type codePatch struct {
	startAddress int
	endAddress   int
	newCode      []byte
}

func rewriteBytecode(originalCode []byte, patches []codePatch) []byte {
	addressMapping := buildAddressMapping(originalCode, patches)

	return rebuildCode(originalCode, patches, addressMapping)
}

func buildAddressMapping(original []byte, patches []codePatch) map[int]int {
	mapping := make(map[int]int, 1024)
	patchIndex := 0
	newAddress := 0

	for oldAddress := 0; oldAddress < len(original); {
		if patchIndex < len(patches) && oldAddress == patches[patchIndex].startAddress {
			mapping[oldAddress] = newAddress
			newAddress += len(patches[patchIndex].newCode)
			oldAddress = patches[patchIndex].endAddress
			patchIndex++
			continue
		}

		op := bytecode.OpCode(original[oldAddress])
		size := GetInstructionSize(op)
		if size <= 0 || oldAddress+size > len(original) {
			break
		}

		mapping[oldAddress] = newAddress
		newAddress += size
		oldAddress += size
	}

	mapping[len(original)] = newAddress
	return mapping
}

func rebuildCode(original []byte, patches []codePatch, addressMapping map[int]int) []byte {
	result := make([]byte, 0, len(original))
	patchIndex := 0

	for ip := 0; ip < len(original); {
		if patchIndex < len(patches) && ip == patches[patchIndex].startAddress {
			result = append(result, patches[patchIndex].newCode...)
			ip = patches[patchIndex].endAddress
			patchIndex++
			continue
		}

		op := bytecode.OpCode(original[ip])
		ip++
		result = append(result, byte(op))

		switch op {
		case bytecode.OpConst, bytecode.OpCall:
			if ip+1 >= len(original) {
				return original
			}
			result = append(result, original[ip], original[ip+1])
			ip += 2

		case bytecode.OpJump, bytecode.OpJumpIfFalse:
			if ip+1 >= len(original) {
				return original
			}
			oldTarget := int(uint16(original[ip])<<8 | uint16(original[ip+1]))
			ip += 2

			newTarget, exists := addressMapping[oldTarget]
			if !exists {
				return original
			}

			result = append(result, byte(uint16(newTarget)>>8), byte(uint16(newTarget)))

		case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
			if ip >= len(original) {
				return original
			}
			result = append(result, original[ip])
			ip++
		}
	}

	return result
}
