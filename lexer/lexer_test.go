package lexer

import (
	"testing"
	tk "github.com/theawakener0/zod/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
	let ten = 10;

	let add = fn(x, y) {
		x+y;
	};

	let result = add(five, ten);
	`

	tests := []tk.Token {
		{tk.LET, "let"},
		{tk.IDENT, "five"},
		{tk.ASSIGN, "="},
		{tk.INT, "5"},
		{tk.SEMICOLON, ";"},
		{tk.LET, "let"},
		{tk.IDENT, "ten"},
		{tk.ASSIGN, "="},
		{tk.INT, "10"},
		{tk.SEMICOLON, ";"},
		{tk.LET, "let"},
		{tk.IDENT, "add"},
		{tk.ASSIGN, "="},
		{tk.FUNCTION, "fn"},
		{tk.LPAREN, "("},
		{tk.IDENT, "x"},
		{tk.COMMA, ","},
		{tk.IDENT, "y"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.IDENT, "x"},
		{tk.PLUS, "+"},
		{tk.IDENT, "y"},
		{tk.SEMICOLON, ";"},
		{tk.RBRACE, "}"},
		{tk.SEMICOLON, ";"},
		{tk.LET, "let"},
		{tk.IDENT, "result"},
		{tk.ASSIGN, "="},
		{tk.IDENT, "add"},
		{tk.LPAREN, "("},
		{tk.IDENT, "five"},
		{tk.COMMA, ","},
		{tk.IDENT, "ten"},
		{tk.RPAREN, ")"},
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



