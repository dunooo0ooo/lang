package bytecode

type FunctionInfo struct {
	Name       string
	ParamCount int
	ParamTypes []TypeKind
	ReturnType TypeKind
	Chunk      Chunk
	NumLocals  int
}

func CreateFunction(name string, paramCount int) *FunctionInfo {
	return &FunctionInfo{
		Name:       name,
		ParamCount: paramCount,
		ParamTypes: make([]TypeKind, 0, paramCount),
		Chunk:      Chunk{},
	}
}

func (f *FunctionInfo) AddParameter(paramType TypeKind) {
	f.ParamTypes = append(f.ParamTypes, paramType)
}

func (f *FunctionInfo) SetReturnType(returnType TypeKind) {
	f.ReturnType = returnType
}

func (f *FunctionInfo) ReserveLocals(count int) {
	f.NumLocals = count
}