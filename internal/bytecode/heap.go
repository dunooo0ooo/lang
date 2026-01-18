package bytecode

type Object struct {
	Mark  bool
	Type  ObjectType
	Next  *Object
	Items []Value
}

type Heap struct {
	Head       *Object
	NumObjects int
	MaxObjects int
}