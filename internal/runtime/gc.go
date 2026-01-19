package runtime

import "github.com/dunooo0ooo/lang/internal/bytecode"

func (vm *VM) gc() {
	vm.markFromRoots()

	vm.sweep()

	if vm.heap.MaxObjects < 8 {
		vm.heap.MaxObjects = 8
	}
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

	switch obj.Type {
	case bytecode.ObjArray:
		for i := range obj.Items {
			vm.markValue(obj.Items[i])
		}
	default:
	}
}

func (vm *VM) sweep() {
	var prev *bytecode.Object
	cur := vm.heap.Head

	for cur != nil {
		if cur.Mark {
			cur.Mark = false
			prev = cur
			cur = cur.Next
			continue
		}

		vm.heap.NumObjects--

		if prev == nil {
			vm.heap.Head = cur.Next
			cur = vm.heap.Head
			continue
		}

		prev.Next = cur.Next
		cur = prev.Next
	}
}
