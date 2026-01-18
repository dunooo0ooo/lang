package bytecode

type TypeKind byte

const (
	TypeInvalid TypeKind = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeString
	TypeChar
	TypeVoid
	TypeNull
	TypeArray
)

type ValueKind byte

const (
	ValInt ValueKind = iota
	ValFloat
	ValBool
	ValString
	ValChar
	ValNull
	ValObject
)

type ObjectType byte

const (
	ObjArray ObjectType = iota
)