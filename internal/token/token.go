package token

type Type int

const (
	// Special
	ILLEGAL Type = iota
	EOF

	// Ident + literals
	IDENT
	INT
	FLOAT
	STRING
	CHAR
	NULL // null literal

	// Keywords
	LET
	FN
	IF
	ELSE
	WHILE
	FOR
	RETURN
	TRUE
	FALSE
	INT_T    // int
	BOOL_T   // bool
	FLOAT_T  // float
	STRING_T // string
	CHAR_T   // char
	VOID_T

	// Operators
	ASSIGN // =
	PLUS   // +
	MINUS  // -
	STAR   // *
	SLASH  // /
	PERCENT

	BANG // !
	AND  // &&
	OR   // ||

	EQ  // ==
	NEQ // !=
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=

	// Delims
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	COMMA     // ,
	SEMICOLON // ;
	COLON     // :
	ARROW     // ->
	LBRACKET  // [
	RBRACKET  // ]
)

type Position struct {
	Offset int // byte offset
	Line   int // 1-based
	Col    int // 1-based
}

type Token struct {
	Type Type
	Lit  string
	Pos  Position
}

var keywords = map[string]Type{
	"let": LET, "fn": FN, "if": IF, "else": ELSE, "while": WHILE, "for": FOR, "return": RETURN,
	"true": TRUE, "false": FALSE,

	"int": INT_T, "bool": BOOL_T,
	"float": FLOAT_T, "string": STRING_T, "char": CHAR_T, "void": VOID_T,

	"null": NULL,
}

func LookupIdent(s string) Type {
	if t, ok := keywords[s]; ok {
		return t
	}
	return IDENT
}

func (t Type) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case LET:
		return "LET"
	case FN:
		return "FN"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case WHILE:
		return "WHILE"
	case FOR:
		return "FOR"
	case RETURN:
		return "RETURN"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case INT_T:
		return "INT_T"
	case BOOL_T:
		return "BOOL_T"
	case ASSIGN:
		return "ASSIGN"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case STAR:
		return "STAR"
	case SLASH:
		return "SLASH"
	case PERCENT:
		return "PERCENT"
	case BANG:
		return "BANG"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case EQ:
		return "EQ"
	case NEQ:
		return "NEQ"
	case LT:
		return "LT"
	case LTE:
		return "LTE"
	case GT:
		return "GT"
	case GTE:
		return "GTE"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case COMMA:
		return "COMMA"
	case SEMICOLON:
		return "SEMICOLON"
	case COLON:
		return "COLON"
	case ARROW:
		return "ARROW"
	case CHAR:
		return "CHAR"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	case FLOAT_T:
		return "FLOAT_T"
	case STRING_T:
		return "STRING_T"
	case CHAR_T:
		return "CHAR_T"
	case STRING:
		return "STRING"
	case FLOAT:
		return "FLOAT"
	default:
		return "Type(?)"
	}
}
