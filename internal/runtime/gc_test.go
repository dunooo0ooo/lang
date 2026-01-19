package runtime

import (
	"testing"

	"github.com/dunooo0ooo/lang/internal/bytecode"
)

func TestGC_DoesNotCrashAndFrees(t *testing.T) {
	mod := &bytecode.Module{Functions: map[string]*bytecode.FunctionInfo{}}
	vm := NewVM(mod, false)

	// маленький порог, чтобы GC точно запускался
	vm.heap.MaxObjects = 8

	// создаём много массивов, но корней не держим => должно чиститься
	for i := 0; i < 500; i++ {
		obj := vm.newObject(bytecode.ObjArray)
		obj.Items = make([]bytecode.Value, 10)

		// имитируем мусор: не кладём в roots/stack/locals
	}

	// после серии GC NumObjects не обязан быть 0 (зависит от момента),
	// но должен быть разумным и не расти бесконечно
	if vm.heap.NumObjects > vm.heap.MaxObjects {
		t.Fatalf("heap leaked: NumObjects=%d > MaxObjects=%d", vm.heap.NumObjects, vm.heap.MaxObjects)
	}
}

func TestGC_RespectsRoots(t *testing.T) {
	mod := &bytecode.Module{Functions: map[string]*bytecode.FunctionInfo{}}
	vm := NewVM(mod, false)
	vm.heap.MaxObjects = 8

	// держим один объект как root через fake stack
	stack := make([]bytecode.Value, 0, 1)
	locals := make([]bytecode.Value, 0)

	vm.roots = append(vm.roots, rootSet{locals: &locals, stack: &stack})
	defer func() { vm.roots = vm.roots[:len(vm.roots)-1] }()

	alive := vm.newObject(bytecode.ObjArray)
	alive.Items = make([]bytecode.Value, 1)
	stack = append(stack, bytecode.Value{Kind: bytecode.ValObject, Obj: alive})

	// создаём мусор, чтобы точно триггернуть GC
	for i := 0; i < 100; i++ {
		o := vm.newObject(bytecode.ObjArray)
		o.Items = make([]bytecode.Value, 5)
	}

	// запускаем GC явно
	vm.gc()

	// alive должен остаться в списке heap
	found := false
	for o := vm.heap.Head; o != nil; o = o.Next {
		if o == alive {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("rooted object was collected")
	}
}
