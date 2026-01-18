package runtime

import "github.com/dunooo0ooo/lang/internal/bytecode"

func (vm *VM) newObject(t bytecode.ObjectType) *bytecode.Object {
	if vm.heap.MaxObjects == 0 {
		vm.heap.MaxObjects = 8
	}

	if vm.heap.NumObjects >= vm.heap.MaxObjects {
		// TODO vm.gc() тут допиливаем гцшник
	}

	obj := &bytecode.Object{
		Type: t,
		Next: vm.heap.Head,
	}
	vm.heap.Head = obj
	vm.heap.NumObjects++

	return obj
}
