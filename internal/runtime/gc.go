package runtime

import "github.com/dunooo0ooo/lang/internal/bytecode"

// gc запускается перед выделением нового объекта (когда превышен лимит).
func (vm *VM) gc() {
	// 1) mark
	vm.markFromRoots()

	// 2) sweep
	vm.sweep()

	// 3) grow threshold (простая стратегия)
	if vm.heap.MaxObjects < 8 {
		vm.heap.MaxObjects = 8
	}
	// если объектов мало, всё равно держим запас
	vm.heap.MaxObjects = vm.heap.NumObjects*2 + 8
}

func (vm *VM) markFromRoots() {
	for _, rs := range vm.roots {
		if rs.locals != nil {
			for i := range *rs.locals {
				vm.markValue((*rs.locals)[i])
			}
		}
		if rs.stack != nil {
			for i := range *rs.stack {
				vm.markValue((*rs.stack)[i])
			}
		}
	}
}

func (vm *VM) markValue(v bytecode.Value) {
	if v.Kind != bytecode.ValObject || v.Obj == nil {
		return
	}
	vm.markObject(v.Obj)
}

func (vm *VM) markObject(obj *bytecode.Object) {
	if obj == nil || obj.Mark {
		return
	}
	obj.Mark = true

	// трассировка ссылок внутри объекта
	switch obj.Type {
	case bytecode.ObjArray:
		for i := range obj.Items {
			vm.markValue(obj.Items[i])
		}
	default:
		// неизвестные типы объектов пока не поддерживаем
	}
}

func (vm *VM) sweep() {
	// проход по односвязному списку с удалением "мертвых" (Mark=false)
	var prev *bytecode.Object
	cur := vm.heap.Head

	for cur != nil {
		if cur.Mark {
			// объект жив, сброс mark для следующего цикла GC
			cur.Mark = false
			prev = cur
			cur = cur.Next
			continue
		}

		// объект мертв: выкидываем из списка
		vm.heap.NumObjects--

		if prev == nil {
			// удаляем head
			vm.heap.Head = cur.Next
			cur = vm.heap.Head
			continue
		}

		prev.Next = cur.Next
		cur = prev.Next
	}
}
