package jit_optimization

import "github.com/dunooo0ooo/lang/internal/bytecode"

func DetectSwapPattern(code []byte, start int) (bool, []byte, int) {
	reader := CodeReader{code: code, ip: start}

	// Чтение arr[j]
	arraySlot, ok := reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok {
		return false, nil, 0
	}
	indexSlot, ok := reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || !reader.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}

	// Чтение arr[j+1]
	temp, ok := reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != arraySlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != indexSlot ||
		!reader.ExpectInstruction(bytecode.OpConst) ||
		!reader.ExpectInstruction(bytecode.OpAdd) ||
		!reader.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}

	// Проверка условия arr[j] > arr[j+1]
	if !reader.ExpectInstruction(bytecode.OpGt) {
		return false, nil, 0
	}
	skipAddress, ok := reader.ExpectArgument(bytecode.OpJumpIfFalse)
	if !ok || !reader.ExpectInstruction(bytecode.OpPop) {
		return false, nil, 0
	}

	// Чтение tmp = arr[j]
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != arraySlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != indexSlot || !reader.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}
	tempSlot, ok := reader.ExpectArgument(bytecode.OpStoreLocal)
	if !ok {
		return false, nil, 0
	}

	// Чтение arr[j] = arr[j+1]
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != arraySlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != indexSlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != arraySlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != indexSlot ||
		!reader.ExpectInstruction(bytecode.OpConst) ||
		!reader.ExpectInstruction(bytecode.OpAdd) ||
		!reader.ExpectInstruction(bytecode.OpArrayGet) ||
		!reader.ExpectInstruction(bytecode.OpArraySet) {
		return false, nil, 0
	}

	// Чтение arr[j+1] = tmp
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != arraySlot {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != indexSlot || !reader.ExpectInstruction(bytecode.OpConst) || !reader.ExpectInstruction(bytecode.OpAdd) {
		return false, nil, 0
	}
	temp, ok = reader.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || temp != tempSlot || !reader.ExpectInstruction(bytecode.OpArraySet) {
		return false, nil, 0
	}

	// Проверка прыжка в конец
	endAddress, ok := reader.ExpectArgument(bytecode.OpJump)
	if !ok {
		return false, nil, 0
	}

	// Валидация паттерна
	if skipAddress < 0 || skipAddress >= len(code) || 
	   bytecode.OpCode(code[skipAddress]) != bytecode.OpPop || 
	   endAddress != skipAddress+1 {
		return false, nil, 0
	}

	// Генерация оптимизированного кода
	optimizedCode := []byte{
		byte(bytecode.OpLoadLocal), byte(arraySlot),
		byte(bytecode.OpLoadLocal), byte(indexSlot),
		byte(bytecode.OpArraySwapJit),
	}

	patternLength := endAddress - start
	if patternLength <= 0 {
		return false, nil, 0
	}
	
	return true, optimizedCode, patternLength
}