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
	lastToken		tk.TokenType
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
	tok := l.nextToken()
	l.lastToken = tok.Type
	return tok
}

func (l *Lexer) nextToken() tk.Token {
	var tok tk.Token

	if l.skipWhitespaceAndComments() && lastTokenCanEndStatement(l.lastToken) {
		return tk.Token{Type: tk.SEMICOLON, Literal: "\n"}
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
		lit, ok := l.readString()
		if !ok {
			tok = tk.Token{Type: tk.ILLEGAL, Literal: "unterminated string literal"}
		} else {
			tok = tk.Token{Type: tk.STRING, Literal: lit}
		}
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
			// L1: empty 0x/0b/0o prefix with 0 digits -> ILLEGAL (hardened)
			if len(tok.Literal) == 2 && (tok.Literal == "0x" || tok.Literal == "0X" || tok.Literal == "0b" || tok.Literal == "0B" || tok.Literal == "0o" || tok.Literal == "0O") {
				tok.Type = tk.ILLEGAL
				return tok
			}
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

	if l.ch == '0' {
		peek := l.peekChar()
		if peek == 'x' || peek == 'X' || peek == 'b' || peek == 'B' || peek == 'o' || peek == 'O' {
			l.readChar()
			l.readChar()
			afterPrefix := l.position
			switch peek {
			case 'x', 'X':
				for isHexDigit(l.ch) {
					l.readChar()
				}
			case 'b', 'B':
				for l.ch == '0' || l.ch == '1' {
					l.readChar()
				}
			case 'o', 'O':
				for '0' <= l.ch && l.ch <= '7' {
					l.readChar()
				}
			}
			// L1: if no valid digit after prefix (l.position == afterPrefix),
			// return prefix literal as-is (e.g. "0x", "0b"); caller will emit ILLEGAL.
			_ = afterPrefix
			return l.input[position:l.position]
		}
	}

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

func (l *Lexer) skipWhitespace() bool {
	crossedNewline := false
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			crossedNewline = true
		}
		l.readChar()
	}
	return crossedNewline
}

func (l *Lexer) skipWhitespaceAndComments() bool {
	crossedNewline := false

	for {
		if l.skipWhitespace() {
			crossedNewline = true
		}

		if !(l.ch == '/' && (l.peekChar() == '/' || l.peekChar() == '*')) {
			break
		}

		if l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
		} else {
			l.readChar()
			l.readChar()
			sawNewline := false
			for l.ch != 0 && !(l.ch == '*' && l.peekChar() == '/') {
				if l.ch == '\n' || l.ch == '\r' {
					sawNewline = true
				}
				l.readChar()
			}
			// L2: unterminated block comment - loop terminates safely on EOF (l.ch == 0);
			// ideally would emit ILLEGAL, but skipWhitespaceAndComments is not token-producing,
			// so we just ensure no infinite loop / panic. Documented as safe.
			if sawNewline {
				crossedNewline = true
			}
			if l.ch == '*' {
				l.readChar()
				l.readChar()
			}
			// L2: if l.ch == 0 here, comment was unterminated (EOF without closing */) - handled as EOF
		}
	}

	return crossedNewline
}

func lastTokenCanEndStatement(t tk.TokenType) bool {
	switch t {
	case tk.IDENT, tk.INT, tk.FLOAT, tk.STRING,
		tk.TRUE, tk.FALSE, tk.NULL,
		tk.RPAREN, tk.RBRACKET, tk.RBRACE,
		tk.INC, tk.DEC,
		tk.RETURN, tk.BREAK, tk.CONTINUE:
		return true
	default:
		return false
	}
}

func (l *Lexer) readString() (string, bool) {
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
			case 'x':
				l.readChar()
				c1 := l.ch
				if c1 == '"' || c1 == 0 {
					sb.WriteString(`\x`)
					continue
				}
				_, hiOK := hexVal(c1)
				if hiOK {
					c2 := l.peekChar()
					_, loOK := hexVal(c2)
					if loOK {
						l.readChar()
						c2 = l.ch
						hi, _ := hexVal(c1)
						lo, _ := hexVal(c2)
						sb.WriteByte(hi*16 + lo)
					} else {
						sb.WriteString(`\x`)
						sb.WriteByte(c1)
					}
				} else {
					sb.WriteString(`\x`)
					continue
				}
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

	if l.ch == 0 {
		return sb.String(), false
	}
	return sb.String(), true
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func hexVal(ch byte) (byte, bool) {
	switch {
	case '0' <= ch && ch <= '9':
		return ch - '0', true
	case 'a' <= ch && ch <= 'f':
		return ch - 'a' + 10, true
	case 'A' <= ch && ch <= 'F':
		return ch - 'A' + 10, true
	}
	return 0, false
}

func isHexDigit(ch byte) bool {
	_, ok := hexVal(ch)
	return ok
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
