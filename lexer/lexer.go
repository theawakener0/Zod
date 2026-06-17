package lexer

import (
	tk "github.com/theawakener0/zod/token"
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
	case ':':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = tk.Token{tk.ASSIGNCHAR, string(ch) + string(l.ch)}
		} else {
			tok = newToken(tk.COLOMN, l.ch)
		}
	case 0:
		tok.Literal = ""
		tok.Type = tk.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = tk.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = tk.INT
			tok.Literal = l.readNumber()
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

	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position

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

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
