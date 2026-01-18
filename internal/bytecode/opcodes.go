package bytecode

type OpCode byte

const (
	OpConst OpCode = iota
	OpLoadLocal
	OpStoreLocal
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow

	OpEq // =
	OpNe // !=
	OpLt // <
	OpLe // <=
	OpGt // >
	OpGe // >=

	OpNeg // -
	OpNot // !

	OpJump
	OpJumpIfFalse
	OpPop

	OpCall
	OpReturn

	OpArrayNew
	OpArrayGet
	OpArraySet

	OpArraySwapJit

	OpPrint
)