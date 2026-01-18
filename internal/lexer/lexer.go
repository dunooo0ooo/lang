package lexer

import "github.com/dunooo0ooo/lang/internal/token"

type Lexer struct {
	src []byte

	i    int
	next int
	ch   byte

	pos token.Position
}

func New(input string) *Lexer {
	l := &Lexer{
		src: []byte(input),
		pos: token.Position{Offset: 0, Line: 1, Col: 1},
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	tokPos := l.pos

	switch l.ch {
	case 0:
		return token.Token{Type: token.EOF, Lit: "", Pos: tokPos}

	case '+':
		l.readChar()
		return token.Token{Type: token.PLUS, Lit: "+", Pos: tokPos}

	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.ARROW, Lit: "->", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.MINUS, Lit: "-", Pos: tokPos}

	case '*':
		l.readChar()
		return token.Token{Type: token.STAR, Lit: "*", Pos: tokPos}

	case '/':
		if l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			l.skipLineComment()
			return l.NextToken()
		}
		if l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			if !l.skipBlockComment() {
				return token.Token{Type: token.ILLEGAL, Lit: "unterminated comment", Pos: tokPos}
			}
			return l.NextToken()
		}
		l.readChar()
		return token.Token{Type: token.SLASH, Lit: "/", Pos: tokPos}

	case '%':
		l.readChar()
		return token.Token{Type: token.PERCENT, Lit: "%", Pos: tokPos}

	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.EQ, Lit: "==", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.ASSIGN, Lit: "=", Pos: tokPos}

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.NEQ, Lit: "!=", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.BANG, Lit: "!", Pos: tokPos}

	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.LTE, Lit: "<=", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.LT, Lit: "<", Pos: tokPos}

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.GTE, Lit: ">=", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.GT, Lit: ">", Pos: tokPos}

	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.AND, Lit: "&&", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.ILLEGAL, Lit: "&", Pos: tokPos}

	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			l.readChar()
			return token.Token{Type: token.OR, Lit: "||", Pos: tokPos}
		}
		l.readChar()
		return token.Token{Type: token.ILLEGAL, Lit: "|", Pos: tokPos}

	case '(':
		l.readChar()
		return token.Token{Type: token.LPAREN, Lit: "(", Pos: tokPos}
	case ')':
		l.readChar()
		return token.Token{Type: token.RPAREN, Lit: ")", Pos: tokPos}
	case '{':
		l.readChar()
		return token.Token{Type: token.LBRACE, Lit: "{", Pos: tokPos}
	case '}':
		l.readChar()
		return token.Token{Type: token.RBRACE, Lit: "}", Pos: tokPos}
	case ',':
		l.readChar()
		return token.Token{Type: token.COMMA, Lit: ",", Pos: tokPos}
	case ';':
		l.readChar()
		return token.Token{Type: token.SEMICOLON, Lit: ";", Pos: tokPos}
	case ':':
		l.readChar()
		return token.Token{Type: token.COLON, Lit: ":", Pos: tokPos}
	case '[':
		l.readChar()
		return token.Token{Type: token.LBRACKET, Lit: "[", Pos: tokPos}
	case ']':
		l.readChar()
		return token.Token{Type: token.RBRACKET, Lit: "]", Pos: tokPos}
	case '"':
		lit, ok := l.readString()
		if !ok {
			l.readChar()
			return token.Token{Type: token.ILLEGAL, Lit: "unterminated string", Pos: tokPos}
		}
		return token.Token{Type: token.STRING, Lit: lit, Pos: tokPos}
	case '\'':
		lit, ok := l.readCharLit()
		if !ok {
			l.readChar()
			return token.Token{Type: token.ILLEGAL, Lit: "bad char literal", Pos: tokPos}
		}
		return token.Token{Type: token.CHAR, Lit: lit, Pos: tokPos}
	}

	if isLetter(l.ch) {
		lit := l.readIdent()
		return token.Token{Type: token.LookupIdent(lit), Lit: lit, Pos: tokPos}
	}
	if isDigit(l.ch) {
		lit, isFloat := l.readNumber()
		if isFloat {
			return token.Token{Type: token.FLOAT, Lit: lit, Pos: tokPos}
		}
		return token.Token{Type: token.INT, Lit: lit, Pos: tokPos}
	}

	ill := string([]byte{l.ch})
	l.readChar()
	return token.Token{Type: token.ILLEGAL, Lit: ill, Pos: tokPos}
}

func (l *Lexer) readChar() {
	if l.next >= len(l.src) {
		l.ch = 0
		l.i = l.next
		return
	}

	l.ch = l.src[l.next]
	l.i = l.next
	l.next++

	if l.ch == '\n' {
		l.pos.Line++
		l.pos.Col = 1
	} else {
		l.pos.Col++
	}
	l.pos.Offset = l.i
}

func (l *Lexer) peekChar() byte {
	if l.next >= len(l.src) {
		return 0
	}
	return l.src[l.next]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdent() string {
	start := l.i
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return string(l.src[start:l.i])
}

func (l *Lexer) readCharLit() (string, bool) {
	l.readChar()
	if l.ch == 0 || l.ch == '\n' || l.ch == '\'' {
		return "", false
	}
	var lit string
	if l.ch == '\\' {
		l.readChar()
		if l.ch == 0 {
			return "", false
		}
		lit = `\` + string([]byte{l.ch})
	} else {
		lit = string([]byte{l.ch})
	}
	l.readChar()
	if l.ch != '\'' {
		return "", false
	}
	l.readChar()
	return lit, true
}

func (l *Lexer) readNumber() (string, bool) {
	start := l.i
	for isDigit(l.ch) {
		l.readChar()
	}
	isFloat := false
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return string(l.src[start:l.i]), isFloat
}

func (l *Lexer) skipLineComment() {
	for l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() bool {
	for {
		if l.ch == 0 {
			return false
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return true
		}
		l.readChar()
	}
}

func (l *Lexer) readString() (string, bool) {

	l.readChar()
	start := l.i
	for {
		if l.ch == 0 || l.ch == '\n' {
			return "", false
		}
		if l.ch == '"' {
			lit := string(l.src[start:l.i])
			l.readChar()
			return lit, true
		}
		if l.ch == '\\' {
			l.readChar()
			if l.ch == 0 {
				return "", false
			}
		}
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
