package compilation

import (
	"github.com/dunooo0ooo/lang/internal/bytecode"
)

type localVar struct {
	name string
	slot int
	typ  bytecode.TypeKind
}

type Compiler struct {
	mod    *bytecode.Module
	fn     *bytecode.FunctionInfo
	locals []localVar

	breakStack    [][]int
	continueStack [][]int
}

func NewCompiler() *Compiler {
	functions := make(map[string]*bytecode.FunctionInfo)
	module := &bytecode.Module{Functions: functions}

	return &Compiler{mod: module}
}

func (c *Compiler) chunk() *bytecode.Chunk {
	return &c.fn.Chunk
}

func (c *Compiler) Module() *bytecode.Module {
	return c.mod
}

func (c *Compiler) addLocal(name string, typ bytecode.TypeKind) int {
	slot := len(c.locals)
	c.locals = append(c.locals, localVar{name: name, slot: slot, typ: typ})

	if c.fn != nil && slot+1 > c.fn.NumLocals {
		c.fn.NumLocals = slot + 1
	}
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].name == name {
			return c.locals[i].slot, true
		}
	}
	return 0, false
}
