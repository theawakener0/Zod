package evaluator

import (
	lx "github.com/theawakener0/zod/lexer"
	obj "github.com/theawakener0/zod/object"
	ps "github.com/theawakener0/zod/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input		string
		expected	any
	} {
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestEvalReturnStatements(t *testing.T) {
	tests := []struct {
		input		string
		expected	int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
			if (10 > 1) {
				if (10 > 1) {
					return 10;
				}

				return 1;
			}

			`, 10},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input		string
		expected	string
	}{
		{
			"5 + true;",
			"unknown infix operator: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"unknown infix operator: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown prefix operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown infix operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown infix operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown infix operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return true + false;
				}
				return 1;
			}
			`,
			"unknown infix operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		err, ok := eval.(*obj.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", eval, eval)
			continue
		}

		if err.Message != tt.expected {
			t.Errorf("wrong error message. got=%q, expected=%q", err.Message, tt.expected)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input		string
		expected	int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestAssignStatements(t *testing.T) {
	tests := []struct {
		input		string
		expected	int64
	}{
		{"a := 5; a;", 5},
		{"a := 5 * 5; a;", 25},
		{"a := 5; b := a; b;", 5},
		{"a := 5; b := a; c := a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	eval := testEval(input)
	fn, ok := eval.(*obj.Function)
	if !ok {
		t.Errorf("object is not Function. got=%T (%+v)", eval, eval)
		return
	}

	if len(fn.Parameters) != 1 {
		t.Errorf("wrong number of parameters. got=%d, expected=%d", len(fn.Parameters), 1)
		return
	}

	if fn.Parameters[0].String() != "x" {
		t.Errorf("wrong parameter name. got=%s, expected=%s", fn.Parameters[0].String(), "x")
		return
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Errorf("wrong body. got=%q, expected=%q", fn.Body.String(), expectedBody)
		return
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input		string
		expected	int64
	} {
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
		fn(y) { x + y };
	};
	let addTwo = newAdder(2);
	addTwo(2);`

	testIntegerObject(t, testEval(input), 4)
}

func testEval(input string) obj.Object {
	l := lx.New(input)
	p := ps.New(l)
	program := p.ParseProgram()
	env := obj.NewEnviroment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, object obj.Object, expected int64) bool {
	result, ok := object.(*obj.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", object, object)
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, expected=%d", result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, object obj.Object, expected bool) bool {
	result, ok := object.(*obj.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", object, object)
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, expected=%t", result.Value, expected)
		return false
	}

	return true
}

func testNullObject(t *testing.T, object obj.Object) bool {
	if object != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", object, object)
		return false
	}

	return true
}

