package bytecode

type Value struct {
	Kind ValueKind
	I    int64
	F    float64
	B    bool
	S    string
	C    byte
	Obj  *Object
}