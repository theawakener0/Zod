package lexer

import (
	"testing"
	tk "github.com/theawakener0/zod/token"
)

func TestNextToken(t *testing.T) {
	input := "=+(){},;"

	tests := []tk.Token {
		{tk.ASSIN, "="},
		{tk.PLUS, "+"},
		{tk.LPAREN, "("},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.RBRACE, "}"},
		{tk.COMMA, ","},
		{tk.SEMICOLON, ";"},
		{tk.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.Type {
			t.Fatalf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.Type, tok.Type)
		}

		if tok.Literal != tt.Literal {
			t.Fatalf("test[%d] - literal wrong. expected=%q, got=%q", i, tt.Literal, tok.Literal)
		}
	}
}



