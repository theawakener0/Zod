package parser

import (
	"testing"
	"fmt"
	"github.com/theawakener0/Zod/ast"
	lx "github.com/theawakener0/Zod/lexer"
)

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"let x = 5;", "x", 5},
		{"let y = 10;", "y", 10},
		{"let foobar = 838383;", "foobar", 838383},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestAssignStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      any
	}{
		{"x := 5;", "x", 5},
		{"y := 10;", "y", 10},
		{"foobar := 838383;", "foobar", 838383},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testAssignStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.AssignStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}
func TestIndexAssignStatement(t *testing.T) {
	tests := []struct {
		input          string
		expectedIndex  any
		expectedValue  any
	}{
		{"x[0] = 10;", 0, 10},
		{"x[0] += 2;", 0, 2},
		{"x[2] *= 3;", 2, 3},
		{"x[2] /= 2;", 2, 2},
		{"x[1] -= 4;", 1, 4},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.AssignStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.AssignStatement, got %T", program.Statements[0])
		}

		idx, ok := stmt.Left.(*ast.IndexExpression)
		if !ok {
			t.Fatalf("stmt.Left is not *ast.IndexExpression, got %T", stmt.Left)
		}

		if !testIdentifier(t, idx.Left, "x") {
			return
		}

		if !testLiteralExpression(t, idx.Index, tt.expectedIndex) {
			return
		}

		if !testLiteralExpression(t, stmt.Value, tt.expectedValue) {
			return
		}
	}
}

func TestIndexAssignStatementWithInfixIndex(t *testing.T) {
	input := "myArray[1 + 1] = 5;"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.AssignStatement, got %T", program.Statements[0])
	}

	idx, ok := stmt.Left.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Left is not *ast.IndexExpression, got %T", stmt.Left)
	}

	if !testIdentifier(t, idx.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, idx.Index, 1, "+", 1) {
		return
	}

	if !testLiteralExpression(t, stmt.Value, 5) {
		return
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue any
	}{
		{"return 5;", 5},
		{"return 10;", 10},
		{"return 993322;", 993322},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt no *ast.ReturnStatement, got %T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Fatalf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
		}
		if !testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
			return
		}
	}
}

func TestBreakContinueStatements(t *testing.T) {
	tests := []struct {
		input        string
		expectedType ast.Statement
		expectedLit  string
	}{
		{"break;", &ast.BreakStatement{}, "break"},
		{"break", &ast.BreakStatement{}, "break"},
		{"continue;", &ast.ContinueStatement{}, "continue"},
		{"continue", &ast.ContinueStatement{}, "continue"},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if _, ok := stmt.(*ast.BreakStatement); ok {
			if _, want := tt.expectedType.(*ast.BreakStatement); !want {
				t.Fatalf("stmt not *ast.BreakStatement, got %T", stmt)
			}
		} else if _, ok := stmt.(*ast.ContinueStatement); ok {
			if _, want := tt.expectedType.(*ast.ContinueStatement); !want {
				t.Fatalf("stmt not *ast.ContinueStatement, got %T", stmt)
			}
		} else {
			t.Fatalf("stmt is neither break nor continue, got %T", stmt)
		}

		if stmt.TokenLiteral() != tt.expectedLit {
			t.Fatalf("stmt.TokenLiteral not %q, got %q", tt.expectedLit, stmt.TokenLiteral())
		}
	}
}

func TestParsingNullLiteral(t *testing.T) {
	input := "null"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", program.Statements[0])
	}

	if _, ok := stmt.Expression.(*ast.NullLiteral); !ok {
		t.Fatalf("exp not *ast.NullLiteral, got %T", stmt.Expression)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    any
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements, got %d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression, got %T", stmt.Expression)
		}
		if exp.Opt != tt.operator {
			t.Fatalf("exp.Operator is not '%s', got %s", tt.operator, exp.Opt)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingPrefixIncDecExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		right    string
	}{
		{"++x;", "++", "x"},
		{"--x;", "--", "x"},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements, got %d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression, got %T", stmt.Expression)
		}
		if exp.Opt != tt.operator {
			t.Fatalf("exp.Opt is not '%s', got %s", tt.operator, exp.Opt)
		}
		if !testIdentifier(t, exp.Right, tt.right) {
			return
		}
	}
}

func TestParsingPostfixIncDecExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		left     string
	}{
		{"x++;", "++", "x"},
		{"x--;", "--", "x"},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements, got %d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PostfixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PostfixExpression, got %T", stmt.Expression)
		}
		if exp.Opt != tt.operator {
			t.Fatalf("exp.Opt is not '%s', got %s", tt.operator, exp.Opt)
		}
		if !testIdentifier(t, exp.Left, tt.left) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  any
		operator   string
		rightValue any
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range infixTests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements, got %d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got %t", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.InfixExpression, got %T", stmt.Expression)
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}

		if !testLiteralExpression(t, exp.Left, tt.leftValue) {
			return
		}

		if !testLiteralExpression(t, exp.Right, tt.rightValue) {
			return
		}
	}

}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input 			string
		expected 		string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"for (;;) { 5; }",
			"for(;;)5",
		},
		{
			"for (x < y) { x; }",
			"for(;(x < y);)x",
		},
		{
			"for (; x < y; ++x) { x; }",
			"for(;(x < y);(++x))x",
		},
		{
			"for (; x; ) { y; }",
			"for(;x;)y",
		},
		{
			"for (i := 0; i < 5; i++) { x; }",
			"for(i := 0;(i < 5);(i++))x",
		},
		{
			"for (i := 0; i < 5; ++i) { x; }",
			"for(i := 0;(i < 5);(++i))x",
		},
		{
			"for (let i = 0; i < 5; i--) { x; }",
			"for(let i = 0;(i < 5);(i--))x",
		},
		{
			"for (x = 5; x > 0; x -= 1) { x; }",
			"for(x = 5;(x > 0);x -= 1)x",
		},
		{
			"for (i := 0; ; i++) { x; }",
			"for(i := 0;;(i++))x",
		},
		{
			"loop { 5; }",
			"loop5",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
		{
			"x[0] = 5",
			"(x[0]) = 5;",
		},
		{
			"x[0] += 5",
			"(x[0]) += 5;",
		},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParseErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct{
		input string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements has %d statements. expected=1", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		boolean, ok := stmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("stmt.Expression is not *ast.Boolean. got=%T", stmt.Expression)
		}

		if boolean.Value != tt.expected {
			t.Errorf("boolean.Value is not %t. got=%t", tt.expected, boolean.Value)
		}

	}
}

func TestIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("exp.Consequence.Statements don't contain %d statements. got=%d\n", 1, len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative is not nil. got=%+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("exp.Consequence.Statements don't contain %d statements. got=%d\n", 1, len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements don't contain %d statements. got=%d\n", 1, len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp.Alternative.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestIfElseIfExpression(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"if (x < y) { x } else if (x > y) { y }"},
		{"if (x < y) { x } elseif (x > y) { y }"},
		{"if (x < y) { x } else if (x > y) { y } else if (x == y) { z }"},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
		}

		if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
			return
		}

		if exp.ElseIf == nil {
			t.Fatalf("exp.ElseIf is nil. got=%+v", exp)
		}

		consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("exp.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
		}
		if !testIdentifier(t, consequence.Expression, "x") {
			return
		}

		next := exp.ElseIf
		if !testInfixExpression(t, next.Condition, "x", ">", "y") {
			return
		}

		elseIfConsequence, ok := next.Consequence.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("exp.ElseIf.Consequence.Statements[0] is not ast.ExpressionStatement. got=%T", next.Consequence.Statements[0])
		}
		if !testIdentifier(t, elseIfConsequence.Expression, "y") {
			return
		}

		for next.ElseIf != nil {
			next = next.ElseIf
		}

		if next.Alternative != nil {
			t.Errorf("deepest ElseIf.Alternative is not nil. got=%+v", next.Alternative)
		}
	}
}

func TestIfElseIfElseExpression(t *testing.T) {
	tests := []struct {
		input               string
		expectedAlternative string
	}{
		{"if (x < y) { x } else if (x > y) { y } else { z }", "z"},
		{"if (x < y) { x } elseif (x > y) { y } else if (x == y) { z } else { w }", "w"},
	}

	for _, tt := range tests {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
		}

		if exp.ElseIf == nil {
			t.Fatalf("exp.ElseIf is nil. got=%+v", exp)
		}

		next := exp.ElseIf
		for next.ElseIf != nil {
			next = next.ElseIf
		}

		if next.Alternative == nil {
			t.Fatalf("deepest ElseIf.Alternative is nil. got=%+v", next)
		}

		if len(next.Alternative.Statements) != 1 {
			t.Errorf("deepest ElseIf.Alternative.Statements don't contain %d statements. got=%d\n", 1, len(next.Alternative.Statements))
			continue
		}

		alternative, ok := next.Alternative.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("deepest ElseIf.Alternative.Statements[0] is not ast.ExpressionStatement. got=%T", next.Alternative.Statements[0])
		}

		if !testIdentifier(t, alternative.Expression, tt.expectedAlternative) {
			continue
		}
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := "fn(x, y) { x + y; }"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(exp.Parameters) != 2 {
		t.Errorf("exp.Parameters don't contain %d parameters. got=%d\n", 2, len(exp.Parameters))
	}

	testLiteralExpression(t, exp.Parameters[0], "x")
	testLiteralExpression(t, exp.Parameters[1], "y")

	bodyStmt, ok := exp.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp.Body.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFuntionParameterParsing(t *testing.T) {
	inputs := []struct {
		input string
		expectedParams []string
	}{
		{
			input: "fn() {};",
			expectedParams: []string{},
		},
		{
			input: "fn(x) {};",
			expectedParams: []string{"x"},
		},
		{
			input: "fn(x, y, z) {};",
			expectedParams: []string{"x", "y", "z"},
		},
	}

	for _, tt := range inputs {
		l := lx.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParseErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		exp := stmt.Expression.(*ast.FunctionLiteral)

		if len(exp.Parameters) != len(tt.expectedParams) {
			t.Errorf("exp.Parameters don't contain %d parameters. got=%d\n", len(tt.expectedParams), len(exp.Parameters))
		}

		for i, param := range tt.expectedParams {
			testLiteralExpression(t, exp.Parameters[i], param)
		}

	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2*3, 4 + 5);"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements don't contain %d statements. got=%d\n", 1, len(program.Statements))
	}
	
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}
	
	if len(exp.Arguments) != 3 {
		t.Fatalf("exp.Arguments don't contain %d arguments. got=%d\n", 3, len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)
	
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value is not 'hello world'. got=%q", literal.Value)
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3];"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("array.Elements don't contain %d elements. got=%d\n", 3, len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1];"

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IndexExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, array.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, array.Index, 1, "+", 1) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Fatalf("hash.Pairs don't contain %d pairs. got=%d\n", 3, len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	
	for k, v := range hash.Pairs {
		literal, ok := k.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("k is not ast.StringLiteral. got=%T", k)
		}

		expectedValue := expected[literal.Value]

		testIntegerLiteral(t, v, expectedValue)
	}
}

func TestParsingEmptyHashLiterals(t *testing.T) {
	input := `{}`

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 0 {
		t.Errorf("hash.Pairs don't contain %d pairs. got=%d\n", 0, len(hash.Pairs))
	}
}

func TestParsingHashLiteralsWithExpressionKeys(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`

	l := lx.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs don't contain %d pairs. got=%d\n", 3, len(hash.Pairs))
	}

	tests := map[string]func(ast.Expression) {
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	
	for k, v := range hash.Pairs {
		literal, ok := k.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("k is not ast.StringLiteral. got=%T", k)
		}

		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("no test for key %q", literal.Value)
			continue
		}
		
		testFunc(v)
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral no 'let', got %q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement, got %T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not %s, got %s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s', got %s", name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testAssignStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != ":=" {
		t.Errorf("s.TokenLiteral no 'let', got %q", s.TokenLiteral())
		return false
	}

	assignStmt, ok := s.(*ast.AssignStatement)
	if !ok {
		t.Errorf("s not *ast.AssignStatement, got %T", s)
		return false
	}

	leftIdent, ok := assignStmt.Left.(*ast.Identifier)
	if !ok {
		t.Errorf("assignStmt.Left is not *ast.Identifier, got %T", assignStmt.Left)
		return false
	}

	if leftIdent.Value != name {
		t.Errorf("assignStmt.Left.Value not %s, got %s", name, leftIdent.Value)
		return false
	}

	if leftIdent.TokenLiteral() != name {
		t.Errorf("assignStmt.Left.TokenLiteral() not '%s', got %s", name, leftIdent.TokenLiteral())
		return false
	}

	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il is not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value is not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral() is not %d. got=%q", value, integ.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value is not %q. got=%q", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() is not %q. got=%q", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected any,
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanExpression(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	left any,
	operator string,
	right any,
) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Opt != operator {
		t.Errorf("opExp.Operator is not '%s'. got=%q", operator, opExp.Opt)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testBooleanExpression(t *testing.T, exp ast.Expression, value bool) bool {
	boolean, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp is not *ast.Boolean. got=%T", exp)
		return false
	}

	if boolean.Value != value {
		t.Errorf("boolean.Value is not %t. got=%t", value, boolean.Value)
		return false
	}

	if boolean.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("boolean.TokenLiteral() is not %t. got=%q", value, boolean.TokenLiteral())
		return false
	}

	return true
}

func checkParseErrors(t *testing.T, p *Parser) {
	errors := p.errors

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parse error: %q", msg)
	}

	t.FailNow()
}


