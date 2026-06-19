package parser

import (
	"testing"
	"github.com/theawakener0/zod/ast"
	lx "github.com/theawakener0/zod/lexer"
)

func TestAssigningStatements(t *testing.T) {
	input := `
	let x = 5;
	y := 10;
	let foo = 838;
	bar := 383;
	`
	l := lx.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}
	
	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foo"},
		{"bar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testAssigningStatements(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testAssigningStatements(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" || s.TokenLiteral() != ":=" {
		t.Errorf("s.TokenLiteral not 'let' or ':='. got=%q", s.TokenLiteral())
		return false
	}

	switch stmt := s.(type) {
	case *ast.LetStatement:
		if stmt.Name.Value != name {
			t.Errorf("LetStatement.Name.Value not '%s'. got=%s", name, stmt.Name.Value)
			return false
		}
		if stmt.Name.TokenLiteral() != name {
			t.Errorf("LetStatement.Name not '%s'. got=%s", name, stmt.Name)
			return false
		}
	case *ast.AssignStatement:
		if stmt.Name.Value != name {
			t.Errorf("AssignStatement.Name.Value not '%s'. got=%s", name, stmt.Name.Value)
			return false
		}
		if stmt.Name.TokenLiteral() != name {
			t.Errorf("AssignStatement.Name not '%s'. got=%s", name, stmt.Name)
			return false
		}
	default:
		t.Errorf("s not *ast.LetStatement or *ast.AssignStatement. got=%T", s)
		return false
	}
	
	return true
}


