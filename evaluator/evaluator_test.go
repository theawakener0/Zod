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
		{"if (1 > 2) { 10 } else if (1 < 2) { 20 }", 20},
		{"if (1 > 2) { 10 } elseif (1 < 2) { 20 }", 20},
		{"if (1 > 2) { 10 } else if (1 > 2) { 20 } else { 30 }", 30},
		{"if (1 < 2) { 10 } else if (1 > 2) { 20 } else { 30 }", 10},
		{"if (1 > 2) { 10 } else if (2 > 3) { 20 } else if (3 > 4) { 30 } else { 40 }", 40},
		{"if (1 > 2) { 10 } elseif (2 < 3) { 20 } else { 30 }", 20},
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
		{
			`"Hello" - "World"`,
			"unknown infix operator: STRING - STRING",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
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

func TestStringLiteral(t *testing.T) {
	input := `"Hello, World!"`

	eval := testEval(input)
	str, ok := eval.(*obj.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", eval, eval)
	}

	if str.Value != "Hello, World!" {
		t.Errorf("object has wrong value. got=%q, expected=%q", str.Value, "Hello, World!")
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	eval := testEval(input)
	str, ok := eval.(*obj.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", eval, eval)
	}

	if str.Value != "Hello World!" {
		t.Errorf("object has wrong value. got=%q, expected=%q", str.Value, "Hello World!")
	}
}

func TestBuiltinFunction(t *testing.T) {
	tests := []struct {
		input		string
		expected	any
	} {
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported. got=INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			errObj, ok := eval.(*obj.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestForExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let x = 5; for (; x < 10; ++x) { }; x;", 10},
		{"let f = fn() { let i = 0; for (; i < 5; ++i) { if (i == 3) { return i; } } }; f()", 3},
		{"let x = 0; for (let i = 0; i < 5; i++) { x = x + i; }; x;", 10},
		{"let x = 0; for (i := 0; i < 5; i++) { x = x + 1; }; x;", 5},
		{"let x = 0; for (x = 0; x < 5; x += 1) { }; x;", 5},
		{"let x = 0; for (x := 0; x < 3; x++) { }; x;", 3},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestForWhileExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let x = 5; for (x < 10) { ++x; }; x;", 10},
		{"let f = fn() { let i = 0; for (i < 3) { if (i == 2) { return i; }; ++i; } }; f()", 2},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestLoopExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let f = fn() { let i = 0; loop { if (i == 5) { return i; }; ++i; } }; f()", 5},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestPrefixIncDec(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let x = 5; ++x; x;", 6},
		{"let x = 5; --x; x;", 4},
		{"let x = 5; ++x; ++x; x;", 7},
		{"let x = 5; --x; --x; x;", 3},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestPostfixIncDec(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let x = 5; x++; x;", 6},
		{"let x = 5; x--; x;", 4},
		{"let x = 5; x++; x++; x;", 7},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestIndexAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"x := [1, 2, 3]; x[0] = 10; x[0];", 10},
		{"x := [1, 2, 3]; x[1] = 7; x[1];", 7},
		{"x := [1, 2, 3]; x[0] = 5; x[2];", 3},
		{"x := [1, 2, 3]; x[0] += 10; x[0];", 11},
		{"x := [1, 2, 3]; x[1] -= 1; x[1];", 1},
		{"x := [2, 2, 2]; x[0] *= 3; x[0];", 6},
		{"x := [10, 2, 3]; x[0] /= 2; x[0];", 5},
		{"x := [1, 2, 3]; x[x[0]] = 9; x[1];", 9},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestIndexAssignmentErrors(t *testing.T) {
	tests := []struct {
		input      string
		expected   string
	}{
		{"x := [1, 2, 3]; x[5] = 10;", "index out of range: 5"},
		{"x := [1, 2, 3]; x[-1] = 10;", "index out of range: -1"},
		{"x := 5; x[0] = 10;", "index assignment requires array or hash, got INTEGER"},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		errObj, ok := eval.(*obj.Error)
		if !ok {
			t.Errorf("eval is not *obj.Error, got %T (%+v)", eval, eval)
			continue
		}
		if errObj.Message != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Message)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	eval := testEval(input)
	array, ok := eval.(*obj.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", eval, eval)
	}

	if len(array.Elements) != 3 {
		t.Errorf("wrong number of elements. got=%d, expected=%d", len(array.Elements), 3)
		return
	}

	testIntegerObject(t, array.Elements[0], 1)
	testIntegerObject(t, array.Elements[1], 4)
	testIntegerObject(t, array.Elements[2], 6)
}

func TestIndexExpression(t *testing.T) {
	tests := []struct {
		input		string
		expected	any
	} {
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
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

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`
	
	eval := testEval(input)
	result, ok := eval.(*obj.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", eval, eval)
	}

	expected := map[obj.HashKey]int64 {
		(&obj.String{Value: "one"}).HashKey(): 1,
		(&obj.String{Value: "two"}).HashKey(): 2,
		(&obj.String{Value: "three"}).HashKey(): 3,
		(&obj.Integer{Value: 4}).HashKey(): 4,
		TRUE.HashKey(): 5,
		FALSE.HashKey(): 6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpression(t *testing.T) {
	tests := []struct {
		input		string
		expected	any
	} {
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
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

func TestHashBuiltins(t *testing.T) {
	tests := []struct {
		input		string
		expected	any
	} {
		{`len({"a": 1, "b": 2})`, 2},
		{`len({})`, 0},
		{`insert({"a": 1}, "b", 2)["b"]`, 2},
		{`len(insert({"a": 1}, "b", 2))`, 2},
		{`insert({"a": 1}, "a", 9)["a"]`, 9},
		{`let h = {"a": 1}; insert(h, "b", 2); len(h)`, 1},
		{`len(delete({"a": 1, "b": 2}, "a"))`, 1},
		{`delete({"a": 1, "b": 2}, "a")["b"]`, 2},
		{`delete({"a": 1, "b": 2}, "c")["a"]`, 1},
		{`len(remove({"a": 1, "b": 2}, "a"))`, 1},
		{`contains({"a": 1}, "a")`, true},
		{`contains({"a": 1}, "b")`, false},
		{`contains({}, "a")`, false},
		{`contains(insert({}, "a", 1), "a")`, true},
		{`len(keys({"a": 1, "b": 2}))`, 2},
		{`len(values({"a": 1, "b": 2}))`, 2},
		{`{"a": 1, "b": 2}["a"]`, 1},
		{`{"a": 1, "b": 2}["c"]`, nil},
		{`let h = {"a": 1}; h["b"] = 2; len(h)`, 2},
		{`let h = {"a": 1}; h["b"] = 2; h["b"]`, 2},
		{`let h = {"a": 1}; h["a"] = 9; h["a"]`, 9},
		{`let h = {"a": 1}; h["a"] += 5; h["a"]`, 6},
		{`let h = {5: 1}; h[5] = 2; h[5]`, 2},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case bool:
			testBooleanObject(t, eval, expected)
		case nil:
			testNullObject(t, eval)
		}
	}
}

func TestHashBuiltinErrors(t *testing.T) {
	tests := []struct {
		input      string
		expected   string
	}{
		{`insert({"a": 1})`, "wrong number of arguments. got=1, want=3"},
		{`insert([1, 2], "a", 1)`, "argument to `insert` not supported. got=ARRAY"},
		{`insert({}, fn(x) { x }, 1)`, "unusable as hash key: FUNCTION"},
		{`delete({"a": 1})`, "wrong number of arguments. got=1, want=2"},
		{`delete(5, "a")`, "argument to `delete` not supported. got=INTEGER"},
		{`keys(1)`, "argument to `keys` not supported. got=INTEGER"},
		{`values("str")`, "argument to `values` not supported. got=STRING"},
		{`contains(1, 1)`, "argument to `contains` not supported. got=INTEGER"},
		{`let h = {}; h["a"] += 1`, "key not found: a"},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		errObj, ok := eval.(*obj.Error)
		if !ok {
			t.Errorf("eval is not *obj.Error, got %T (%+v)", eval, eval)
			continue
		}
		if errObj.Message != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Message)
		}
	}
}

func TestHashKeysValues(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	eval := testEval(`keys(` + input + `)`)
	keys, ok := eval.(*obj.Array)
	if !ok {
		t.Fatalf("keys() did not return Array. got=%T (%+v)", eval, eval)
	}

	if len(keys.Elements) != 3 {
		t.Fatalf("keys() has wrong number of elements. got=%d", len(keys.Elements))
	}

	seen := map[string]bool{}
	for _, el := range keys.Elements {
		str, ok := el.(*obj.String)
		if !ok {
			t.Fatalf("key is not String. got=%T (%+v)", el, el)
		}
		seen[str.Value] = true
	}

	for _, k := range []string{"one", "two", "three"} {
		if !seen[k] {
			t.Errorf("keys() missing key %q", k)
		}
	}

	eval = testEval(`values(` + input + `)`)
	values, ok := eval.(*obj.Array)
	if !ok {
		t.Fatalf("values() did not return Array. got=%T (%+v)", eval, eval)
	}

	if len(values.Elements) != 3 {
		t.Fatalf("values() has wrong number of elements. got=%d", len(values.Elements))
	}

	seenVals := map[int64]bool{}
	for _, el := range values.Elements {
		integer, ok := el.(*obj.Integer)
		if !ok {
			t.Fatalf("value is not Integer. got=%T (%+v)", el, el)
		}
		seenVals[integer.Value] = true
	}

	for _, v := range []int64{1, 2, 3} {
		if !seenVals[v] {
			t.Errorf("values() missing value %d", v)
		}
	}
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

