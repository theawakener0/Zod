package lexer

import (
	"strings"

	tk "github.com/theawakener0/Zod/token"
)

type Lexer struct {
	input			string
	position		int
	readPosition	int
	ch 				byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition

	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func newToken(tokenType tk.TokenType, ch byte) tk.Token {
	return tk.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) NextToken() tk.Token {
	var tok tk.Token

	l.skipWhitespace()

	for l.ch == '/' && (l.peekChar() == '/' || l.peekChar() == '*') {
		if l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
		} else {
			l.readChar()
			l.readChar()
			for l.ch != 0 && !(l.ch == '*' && l.peekChar() == '/') {
				l.readChar()
			}
			if l.ch == '*' {
				l.readChar()
				l.readChar()
			}
		}
		l.skipWhitespace()
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.EQ, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.ASSIGN, l.ch)
		}
	case ';':
		tok = newToken(tk.SEMICOLON, l.ch)
	case '(':
		tok = newToken(tk.LPAREN, l.ch)
	case ')':
		tok = newToken(tk.RPAREN, l.ch)
	case ',':
		tok = newToken(tk.COMMA, l.ch)
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.INCASSIGN, string(ch) + string(l.ch)}
		} else if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.INC, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.DECDASSIGN, string(ch) + string(l.ch)}
		} else if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.DEC, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.MINUS, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.MLTASSIGN, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.ASTERISK, l.ch)
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.DIVASSIGN, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.SLASH, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.NOTEQ, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.BANG, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.LTEQ, string(ch) + string(l.ch)}
		} else {
			tok	= newToken(tk.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.GTEQ, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.GT, l.ch)
		}
	case '{':
		tok = newToken(tk.LBRACE, l.ch)
	case '}':
		tok = newToken(tk.RBRACE, l.ch)
	case '[':
		tok = newToken(tk.LBRACKET, l.ch)
	case ']':
		tok = newToken(tk.RBRACKET, l.ch)
	case ':':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.ASSIGNCHAR, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.COLOMN, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.LAND, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.LOR, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.ILLEGAL, l.ch)
		}
	case '"':
		tok.Type = tk.STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = tk.EOF
	case '.':
		if isDigit(l.peekChar()) {
			tok.Type = tk.FLOAT
			tok.Literal = l.readDotNumber()
			return tok
		}
		tok = newToken(tk.DOT, l.ch)
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = tk.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			if strings.Contains(tok.Literal, ".") {
				tok.Type = tk.FLOAT
			} else {
				tok.Type = tk.INT
			}
			return tok
		} else {
			tok = newToken(tk.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position

	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) readDotNumber() string {
	position := l.position

	l.readChar()

	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	l.readChar()

	var sb strings.Builder
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case 0:
				sb.WriteByte('\\')
			default:
				sb.WriteByte('\\')
				sb.WriteByte(l.ch)
			}
			l.readChar()
			continue
		}
		sb.WriteByte(l.ch)
		l.readChar()
	}

	return sb.String()
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
