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
	20.0
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
		{tk.SEMICOLON, "\n"},
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
		{tk.SEMICOLON, "\n"},
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
		{tk.SEMICOLON, "\n"},
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
		{tk.SEMICOLON, "\n"},
		{tk.STRING, "foo bar"},
		{tk.SEMICOLON, "\n"},
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
		{tk.SEMICOLON, "\n"},
		{tk.FLOAT, "20.0"},
		{tk.SEMICOLON, "\n"},
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

func TestFloatTokens(t *testing.T) {
	input := "20.0; 3.14; 20.; .5; 5;"

	tests := []tk.Token{
		{tk.FLOAT, "20.0"},
		{tk.SEMICOLON, ";"},
		{tk.FLOAT, "3.14"},
		{tk.SEMICOLON, ";"},
		{tk.FLOAT, "20."},
		{tk.SEMICOLON, ";"},
		{tk.FLOAT, ".5"},
		{tk.SEMICOLON, ";"},
		{tk.INT, "5"},
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

func TestHexEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ansi", `"\x1b[0;34m%s\x1b[0m"`, "\x1b[0;34m%s\x1b[0m"},
		{"uppercase", `"\x41\x42"`, "AB"},
		{"invalid", `"\xZZ"`, `\xZZ`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tk.STRING {
				t.Fatalf("expected STRING token, got %q", tok.Type)
			}
			if tok.Literal != tt.want {
				t.Fatalf("literal wrong. expected=%q, got=%q", tt.want, tok.Literal)
			}
		})
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

func TestNewlineSemicolonInsertion(t *testing.T) {
	input := `let a = 5
let b = 6
println(a)
x := 1 +
	2
y := [1,
	2]
if (a > 5)
{
	z := "s" + "t"
}
`

	tests := []tk.Token{
		{tk.LET, "let"},
		{tk.IDENT, "a"},
		{tk.ASSIGN, "="},
		{tk.INT, "5"},
		{tk.SEMICOLON, "\n"},
		{tk.LET, "let"},
		{tk.IDENT, "b"},
		{tk.ASSIGN, "="},
		{tk.INT, "6"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "println"},
		{tk.LPAREN, "("},
		{tk.IDENT, "a"},
		{tk.RPAREN, ")"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "x"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "1"},
		{tk.PLUS, "+"},
		{tk.INT, "2"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "y"},
		{tk.ASSIGNCHAR, ":="},
		{tk.LBRACKET, "["},
		{tk.INT, "1"},
		{tk.COMMA, ","},
		{tk.INT, "2"},
		{tk.RBRACKET, "]"},
		{tk.SEMICOLON, "\n"},
		{tk.IF, "if"},
		{tk.LPAREN, "("},
		{tk.IDENT, "a"},
		{tk.GT, ">"},
		{tk.INT, "5"},
		{tk.RPAREN, ")"},
		{tk.SEMICOLON, "\n"},
		{tk.LBRACE, "{"},
		{tk.IDENT, "z"},
		{tk.ASSIGNCHAR, ":="},
		{tk.STRING, "s"},
		{tk.PLUS, "+"},
		{tk.STRING, "t"},
		{tk.SEMICOLON, "\n"},
		{tk.RBRACE, "}"},
		{tk.SEMICOLON, "\n"},
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

func TestNoSemicolonAfterOperator(t *testing.T) {
	input := `x := 1 +
2
y := 3
a :=
4
`

	tests := []tk.Token{
		{tk.IDENT, "x"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "1"},
		{tk.PLUS, "+"},
		{tk.INT, "2"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "y"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "3"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "a"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "4"},
		{tk.SEMICOLON, "\n"},
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

func TestNoDoubleSemicolonOnBlankLines(t *testing.T) {
	input := `x := 1

y := 2
`

	tests := []tk.Token{
		{tk.IDENT, "x"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "1"},
		{tk.SEMICOLON, "\n"},
		{tk.IDENT, "y"},
		{tk.ASSIGNCHAR, ":="},
		{tk.INT, "2"},
		{tk.SEMICOLON, "\n"},
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

func TestElseIfKeywords(t *testing.T) {
	input := "if (x) { } else if (y) { } elseif (z) { }"

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



