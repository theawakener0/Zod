package ast

import (
	tk "github.com/theawakener0/zod/token"
	"testing"
)


func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: tk.Token{Type: tk.LET, Literal: "let"},
				Name: &Identifier{
					Token: tk.Token{Type: tk.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: tk.Token{Type: tk.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if program.String() != "let myVar = anotherVar;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}


