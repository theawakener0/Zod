package lexer

import (
	"testing"
	tk "github.com/theawakener0/Zod/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
	let ten = 10;

	let add = fn(x, y) {
		x+y;
	};

	FiveAndTen := 10 + 5 * 1;

	loop {FiveAndTen -= 1;}

	if (five >= ten) {
		for (;;) {
			FiveAndTen += 1/1;
		}
	} else {
		end := add(FiveAndTen, ten) / 2; 
	}


	let result = add(five, ten);
	"foobar"
	"foo bar"
	[1, 2];
	{"foo": "bar"}
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
		{tk.IDENT, "FiveAndTen"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "10"},
		{tk.PLUS, "+"},
		{tk.INT, "5"},
		{tk.ASTERISK, "*"},
		{tk.INT, "1"},
		{tk.SEMICOLON, ";"},
		{tk.LOOP, "loop"},
		{tk.LBRACE, "{"},
		{tk.IDENT, "FiveAndTen"},
		{tk.DECDASSIGN, "-="},
		{tk.INT, "1"},
		{tk.SEMICOLON, ";"},
		{tk.RBRACE, "}"},
		{tk.IF, "if"},
		{tk.LPAREN, "("},
		{tk.IDENT, "five"},
		{tk.GTEQ, ">="},
		{tk.IDENT, "ten"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.FOR, "for"},
		{tk.LPAREN, "("},
		{tk.SEMICOLON, ";"},
		{tk.SEMICOLON, ";"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.IDENT, "FiveAndTen"},
		{tk.INCASSIGN, "+="},
		{tk.INT, "1"},
		{tk.SLASH, "/"},
		{tk.INT, "1"},
		{tk.SEMICOLON, ";"},
		{tk.RBRACE, "}"},
		{tk.RBRACE, "}"},
		{tk.ELSE, "else"},
		{tk.LBRACE, "{"},
		{tk.IDENT, "end"},
		{tk.ASSIGNCHAR, ":="},
		{tk.IDENT, "add"},
		{tk.LPAREN, "("},
		{tk.IDENT, "FiveAndTen"},
		{tk.COMMA, ","},
		{tk.IDENT, "ten"},
		{tk.RPAREN, ")"},
		{tk.SLASH, "/"},
		{tk.INT, "2"},
		{tk.SEMICOLON, ";"},
		{tk.RBRACE, "}"},
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
		{tk.STRING, "foobar"},
		{tk.STRING, "foo bar"},
		{tk.LBRACKET, "["},
		{tk.INT, "1"},
		{tk.COMMA, ","},
		{tk.INT, "2"},
		{tk.RBRACKET, "]"},
		{tk.SEMICOLON, ";"},
		{tk.LBRACE, "{"},
		{tk.STRING, "foo"},
		{tk.COLOMN, ":"},
		{tk.STRING, "bar"},
		{tk.RBRACE, "}"},
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

func TestStringEscapes(t *testing.T) {
	input := `"hello\nworld\t\"quoted\"\\done"`

	l := New(input)
	tok := l.NextToken()

	if tok.Type != tk.STRING {
		t.Fatalf("expected STRING token, got %q", tok.Type)
	}
	if tok.Literal != "hello\nworld\t\"quoted\"\\done" {
		t.Fatalf("literal wrong. expected=%q, got=%q", "hello\nworld\t\"quoted\"\\done", tok.Literal)
	}
}

func TestComments(t *testing.T) {
	input := `5; // line comment
	// full line comment
	let x = 5; /* block */ x;
	/* multi
	line */ 7;
	x = 10 / 2; // trailing
	/* unterminated
	`

	tests := []tk.Token{
		{Type: tk.INT, Literal: "5"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.LET, Literal: "let"},
		{Type: tk.IDENT, Literal: "x"},
		{Type: tk.ASSIGN, Literal: "="},
		{Type: tk.INT, Literal: "5"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.IDENT, Literal: "x"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.INT, Literal: "7"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.IDENT, Literal: "x"},
		{Type: tk.ASSIGN, Literal: "="},
		{Type: tk.INT, Literal: "10"},
		{Type: tk.SLASH, Literal: "/"},
		{Type: tk.INT, Literal: "2"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.EOF, Literal: ""},
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

func TestNewKeywords(t *testing.T) {
	input := "break; continue; null"

	tests := []tk.Token{
		{Type: tk.BREAK, Literal: "break"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.CONTINUE, Literal: "continue"},
		{Type: tk.SEMICOLON, Literal: ";"},
		{Type: tk.NULL, Literal: "null"},
		{Type: tk.EOF, Literal: ""},
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

func TestElseIfKeywords(t *testing.T) {	input := "if (x) { } else if (y) { } elseif (z) { }"

	tests := []tk.Token{
		{tk.IF, "if"},
		{tk.LPAREN, "("},
		{tk.IDENT, "x"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.RBRACE, "}"},
		{tk.ELSE, "else"},
		{tk.IF, "if"},
		{tk.LPAREN, "("},
		{tk.IDENT, "y"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.RBRACE, "}"},
		{tk.ELSEIF, "elseif"},
		{tk.LPAREN, "("},
		{tk.IDENT, "z"},
		{tk.RPAREN, ")"},
		{tk.LBRACE, "{"},
		{tk.RBRACE, "}"},
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



