package bytecode

type Module struct {
	Name      string
	Functions map[string]*FunctionInfo
}

func CreateModule(name string) *Module {
	return &Module{
		Name:      name,
		Functions: make(map[string]*FunctionInfo),
	}
}

func (m *Module) AddFunction(fn *FunctionInfo) error {
	if _, exists := m.Functions[fn.Name]; exists {
		return &DuplicateFunctionError{Name: fn.Name}
	}
	m.Functions[fn.Name] = fn
	return nil
}

func (m *Module) GetFunction(name string) (*FunctionInfo, bool) {
	fn, exists := m.Functions[name]
	return fn, exists
}

func (m *Module) FunctionCount() int {
	return len(m.Functions)
}

type DuplicateFunctionError struct {
	Name string
}

func (e *DuplicateFunctionError) Error() string {
	return "duplicate function: " + e.Name
}