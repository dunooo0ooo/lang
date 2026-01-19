package sema

import "github.com/dunooo0ooo/lang/internal/bytecode"

type Type struct {
	Kind bytecode.TypeKind
	Elem *Type
}

func T(k bytecode.TypeKind) Type { return Type{Kind: k} }

func Arr(elem Type) Type {
	e := elem
	return Type{Kind: bytecode.TypeArray, Elem: &e}
}

func (t Type) IsArray() bool { return t.Kind == bytecode.TypeArray }

func (t Type) Equal(u Type) bool {
	if t.Kind != u.Kind {
		return false
	}
	if t.Kind != bytecode.TypeArray {
		return true
	}
	if t.Elem == nil || u.Elem == nil {
		return t.Elem == u.Elem
	}
	return t.Elem.Equal(*u.Elem)
}

func (t Type) String() string {
	switch t.Kind {
	case bytecode.TypeInt:
		return "int"
	case bytecode.TypeFloat:
		return "float"
	case bytecode.TypeBool:
		return "bool"
	case bytecode.TypeString:
		return "string"
	case bytecode.TypeChar:
		return "char"
	case bytecode.TypeVoid:
		return "void"
	case bytecode.TypeNull:
		return "null"
	case bytecode.TypeArray:
		if t.Elem == nil {
			return "[]<?>"
		}
		return "[]" + t.Elem.String()
	default:
		return "<?>"
	}
}

func IsRefType(t Type) bool {
	return t.Kind == bytecode.TypeString || t.Kind == bytecode.TypeArray
}
