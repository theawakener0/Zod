package vm

import (
	"fmt"
	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/compiler"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"testing"
)

type vmTestCase struct {
	input    string
	expected any
}

func TestIntegerArthimetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVMTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
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
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	runVMTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"!(if (false) { 5; })", true},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVMTests(t, tests)
}

func TestGlobalAssigningStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVMTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"hello"`, "hello"},
		{`"hello" + " world!"`, "hello world!"},
		{`"Zod" + " Programming" + " Language"`, "Zod Programming Language"},
	}

	runVMTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVMTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[obj.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[obj.HashKey]int64{
				(&obj.Integer{Value: 1}).HashKey(): 2,
				(&obj.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[obj.HashKey]int64{
				(&obj.Integer{Value: 2}).HashKey(): 4,
				(&obj.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVMTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "fivePlusTen := fn() { 5 + 10 }; fivePlusTen();",
			expected: 15,
		},
		{
			input:    "one := fn() { 1 }; two := fn() { 2 }; one() + two()",
			expected: 3,
		},
		{
			input:    "let a = fn() { 1 }; b := fn() { a() + 1}; let c = fn() { b() + 1 }; c();",
			expected: 3,
		},
	}

	runVMTests(t, tests)
}

func TestFunctionWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let earlyExit = fn() {return 99; 100}; earlyExit();",
			expected: 99,
		},
		{
			input:    "let earlyExit = fn() {return 99; return 100}; earlyExit();",
			expected: 99,
		},
	}

	runVMTests(t, tests)
}

func TestFunctionWithoutReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let noReturn = fn() { }; noReturn()",
			expected: Null,
		},
		{
			input:    "let noReturn = fn() { }; noReturnTwo := fn() { noReturn() }; noReturnTwo(); noReturn()",
			expected: Null,
		},
	}

	runVMTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let returnsOne = fn() { 1 }; returnsOneReturner := fn() { returnsOne }; returnsOneReturner()();",
			expected: 1,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let one = fn() { one := 1; one }; one();",
			expected: 1,
		},
		{
			input:    "let oneAndTwo = fn() { let one = 1; two := 2; one + two }; oneAndTwo();",
			expected: 3,
		},
		{
			input:    "let oneAndTwo = fn() { let one = 1; let two = 2; one + two; }; let threeAndFour = fn() { let three = 3; let four = 4; three + four; }; oneAndTwo() + threeAndFour();",
			expected: 10,
		},
		{
			input:    "let firstFoobar = fn() { let foobar = 50; foobar; }; let secondFoobar = fn() { let foobar = 100; foobar; }; firstFoobar() + secondFoobar();",
			expected: 150,
		},
		{
			input:    "let globalSeed = 50; let minusOne = fn() { let num = 1; globalSeed - num; } let minusTwo = fn() { let num = 2; globalSeed - num; } minusOne() + minusTwo();",
			expected: 97,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let identity = fn(a) { a; }; identity(4)",
			expected: 4,
		},
		{
			input:    "sum := fn(a, b) { a + b }; sum(1, 2)",
			expected: 3,
		},
		{
			input:    "sum := fn(a, b) { let c = a + b; c }; sum(1, 2);",
			expected: 3,
		},
		{
			input:    "let sum = fn(a, b) { c := a + b; c }; sum(1, 2) + sum(3, 4);",
			expected: 10,
		},
		{
			input:    "sum := fn(a, b) { let c = a + b; c }; outer := fn() { sum(1, 2) + sum(3, 4) }; outer()",
			expected: 10,
		},
	}

	runVMTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "fn() { 1 }(1);",
			expected: "wrong number of arguments. got=1, want=0",
		},
		{
			input:    "fn(a) { a }();",
			expected: "wrong number of arguments. got=0, want=1",
		},
		{
			input:    "fn(a, b) { a + b }(1);",
			expected: "wrong number of arguments. got=1, want=2",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		errObj, ok := stackElem.(*obj.Error)
		if !ok {
			t.Fatalf("object is not Error. got=%T (%+v)", stackElem, stackElem)
		}

		if errObj.Message != tt.expected {
			t.Errorf("wrong VM error: want=%q, got=%q", tt.expected, errObj.Message)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `len("")`,
			expected: 0,
		},
		{
			input:    `len("four")`,
			expected: 4,
		},
		{
			input:    `len("hello world")`,
			expected: 11,
		},
		{
			input: `len(1)`,
			expected: &obj.Error{
				Message: "argument to `len` not supported. got=INTEGER",
			},
		},
		{
			input: `len("one", "two")`,
			expected: &obj.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{
			input:    `len([1, 2, 3])`,
			expected: 3,
		},
		{
			input:    `len([])`,
			expected: 0,
		},
		{
			input:    `println("hello", "world!")`,
			expected: Null,
		},
		{
			input:    `first([1, 2, 3])`,
			expected: 1,
		},
		{
			input:    `first([])`,
			expected: Null,
		},
		{
			input: `first(1)`,
			expected: &obj.Error{
				Message: "argument to `first` not supported. got=INTEGER",
			},
		},
		{
			input:    `last([1, 2, 3])`,
			expected: 3,
		},
		{
			input:    `last([])`,
			expected: Null,
		},
		{
			input: `last(1)`,
			expected: &obj.Error{
				Message: "argument to `last` not supported. got=INTEGER",
			},
		},
		{
			input:    `pop([1, 2, 3])`,
			expected: []int{1, 2},
		},
		{
			input:    `pop([])`,
			expected: Null,
		},
		{
			input:    `push([], 1)`,
			expected: []int{1},
		},
		{
			input: `push(1, 1)`,
			expected: &obj.Error{
				Message: "argument to `push` not supported. got=INTEGER",
			},
		},
	}

	runVMTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "newClosure := fn(a) { fn() { a } }; let closure = newClosure(99); closure();",
			expected: 99,
		},
	}

	runVMTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "countDown := fn(x) { if (x == 0) { return 0; } else { countDown(x - 1) } }; countDown(1)",
			expected: 0,
		},
		{
			input:    "countDown := fn(x) { if (x == 0) { return 0; } else { countDown(x - 1) } }; let wrapper = fn() { countDown(1) }; wrapper()",
			expected: 0,
		},
		{
			input:    "let wrapper = fn() { countDown := fn(x) { if (x == 0) { return 0; } else { countDown(x - 1) } }; countDown(1) }; wrapper()",
			expected: 0,
		},
	}

	runVMTests(t, tests)
}

func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "fibonacci := fn(x) { if (x==0) { return 0; } else { if (x==1) { return 1; } else { fibonacci(x-1) + fibonacci(x-2) } } }; fibonacci(15)",
			expected: 610,
		},
	}

	runVMTests(t, tests)
}

func runVMTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err0 := comp.Compile(program)
		if err0 != nil {
			t.Fatalf("compiler error: %s", err0)
		}

		for i, constant := range comp.Bytecode().Constant {
			fmt.Printf("CONSTANT %d %p (%T):\n", i, constant, constant)

			switch constant := constant.(type) {
			case *obj.CompiledFunction:
				fmt.Printf(" Instructions:\n%s", constant.Instructions)
			case *obj.Integer:
				fmt.Printf(" Value: %d\n", constant.Value)
			}
		}

		vm := New(comp.Bytecode())
		err1 := vm.Run()
		if err1 != nil {
			t.Fatalf("vm error: %s", err1)
		}

		stackElem := vm.LastPoppedStackElem()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(
	t *testing.T,
	expected any,
	actual obj.Object,
) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case string:
		err := testStringObject(string(expected), actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case []int:
		arr, ok := actual.(*obj.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", actual, actual)
		}

		if len(arr.Elements) != len(expected) {
			t.Errorf("object has wrong length. got=%d, want=%d", len(arr.Elements), len(expected))
		}

		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), arr.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[obj.HashKey]int64:
		hash, ok := actual.(*obj.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *obj.Error:
		errObj, ok := actual.(*obj.Error)
		if !ok {
			t.Errorf("object is not Error. got=%T (%+v)", actual, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errObj.Message)
		}
	case *obj.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	}
}

func parse(input string) *ast.Program {
	l := lx.New(input)
	p := ps.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual obj.Object) error {
	result, ok := actual.(*obj.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual obj.Object) error {
	result, ok := actual.(*obj.String)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual obj.Object) error {
	result, ok := actual.(*obj.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
	}

	return nil
}

func TestLogicalOperatorsVM(t *testing.T) {
	tests := []vmTestCase{
		{"true && false", false},
		{"true && 5", 5},
		{"false && 5", false},
		{"true || false", true},
		{"false || 5", 5},
		{"false || false", false},
		{"1 < 2 && 3 > 2", true},
	}

	runVMTests(t, tests)
}

func TestCompoundAssignVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 5; x += 2; x;", 7},
		{"let x = 5; x -= 2; x;", 3},
		{"let x = 5; x *= 2; x;", 10},
		{"let x = 5; x /= 2; x;", 2},
		{"let x = 5; let y = ++x; y;", 6},
		{"let x = 5; let y = x++; y;", 5},
		{"-3.14", &obj.Error{Message: ""}},
	}

	tests = tests[:len(tests)-1]
	runVMTests(t, tests)
}

func TestFloatNegationVM(t *testing.T) {
	program := parse("-3.14")
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	machine := New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		t.Fatalf("vm error: %s", err)
	}
	f, ok := machine.LastPoppedStackElem().(*obj.Float)
	if !ok {
		t.Fatalf("object is not Float. got=%T", machine.LastPoppedStackElem())
	}
	if f.Value != -3.14 {
		t.Fatalf("wrong value. got=%v", f.Value)
	}
}

func TestIndexAssignVM(t *testing.T) {
	tests := []vmTestCase{
		{"x := [1, 2, 3]; x[0] = 10; x[0];", 10},
		{"x := [1, 2, 3]; x[0] += 10; x[0];", 11},
		{"x := [1, 2, 3]; x[1] -= 1; x[1];", 1},
		{`let h = {"a": 1}; h["b"] = 2; h["b"]`, 2},
		{`let h = {"a": 1}; h["a"] += 5; h["a"]`, 6},
		{"x := [1, 2, 3]; x[5] = 10;", &obj.Error{Message: "index out of range: 5"}},
		{"10 / 0", &obj.Error{Message: "division by zero"}},
		{"1 + true", &obj.Error{Message: "unknown infix operator: INTEGER + BOOLEAN"}},
	}

	runVMTests(t, tests)
}

func TestForLoopVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 0; for (i := 0; i < 5; i++) { x = x + 1; }; x;", 5},
		{"let x = 5; for (; x < 10; ++x) { }; x;", 10},
		{"let x = 5; for (x < 10) { ++x; }; x;", 10},
		{"let x = 0; for (i := 0; i < 10; i++) { if (i == 5) { break; } x = x + 1; }; x;", 5},
		{"let x = 0; for (i := 0; i < 5; i++) { if (i == 2) { continue; } x = x + 1; }; x;", 4},
		{"let i = 0; loop { if (i == 3) { break; } i++; }; i;", 3},
		{"let x = 0; for (x := 0; x < 3; x++) { }; x;", 0},
	}

	runVMTests(t, tests)
}

func TestTryVM(t *testing.T) {
	program := parse(`try(10 / 0)`)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	machine := New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		t.Fatalf("vm error: %s", err)
	}
	arr, ok := machine.LastPoppedStackElem().(*obj.Array)
	if !ok {
		t.Fatalf("try did not return Array. got=%T", machine.LastPoppedStackElem())
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("try array wrong length. got=%d", len(arr.Elements))
	}
	if b, ok := arr.Elements[0].(*obj.Boolean); !ok || b.Value != false {
		t.Fatalf("try ok flag wrong. got=%v", arr.Elements[0])
	}
	if s, ok := arr.Elements[1].(*obj.String); !ok || s.Value != "division by zero" {
		t.Fatalf("try message wrong. got=%v", arr.Elements[1])
	}
}

func TestMatrixVM(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`matrix(2, 2, [1, 2, 3, 4]) + matrix(2, 2, [1, 1, 1, 1])`, "[[2, 3], [4, 5]]"},
		{`matrix(2, 2, [1, 2, 3, 4]) * 2`, "[[2, 4], [6, 8]]"},
		{`let m = matrix(2, 3, [1, 2, 3, 4, 5, 6]); m[0][1]`, "2"},
		{`let m = matrix(2, 2, [1, 2, 3, 4]); m[0][0] = 9; m[0][0]`, "9"},
		{`let m = matrix(2, 2, [1, 2, 3, 4]); m[0][0] += 10; m[0][0]`, "11"},
	}
	for _, tt := range cases {
		program := parse(tt.input)
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error for %q: %s", tt.input, err)
		}
		machine := New(comp.Bytecode())
		if err := machine.Run(); err != nil {
			t.Fatalf("vm error for %q: %s", tt.input, err)
		}
		if got := machine.LastPoppedStackElem().Inspect(); got != tt.expected {
			t.Errorf("wrong matrix result for %q. want=%q got=%q", tt.input, tt.expected, got)
		}
	}
}

func TestBlockScopingVM(t *testing.T) {
	program := parse(`if (true) { let hidden = 42 }; hidden`)
	comp := compiler.New()
	err := comp.Compile(program)
	if err == nil {
		t.Fatalf("expected compile error for leaked block var, got none")
	}
	if err.Error() != "identifier not found: hidden" {
		t.Fatalf("wrong error message. got=%q", err.Error())
	}
}

func TestSharedFreeMutationVM(t *testing.T) {
	tests := []vmTestCase{
		{"let mk = fn() { let c = 0; let a = fn() { c = c + 1; c }; let b = fn() { c = c + 10; c }; a(); b(); c }; mk()", 11},
		{"let mk = fn() { let c = 0; let bump = fn() { c++ }; bump(); bump(); c }; mk()", 2},
		{"let mk = fn() { let c = 0; let bump = fn() { c += 5 }; bump(); bump(); c }; mk()", 10},
	}

	runVMTests(t, tests)
}

func TestElseIfChainVM(t *testing.T) {
	tests := []vmTestCase{
		{`let g = fn(s) { if (s >= 90) { "A" } elseif (s >= 80) { "B" } else if (s >= 70) { "C" } else { "F" } }; g(95)`, "A"},
		{`let g = fn(s) { if (s >= 90) { "A" } elseif (s >= 80) { "B" } else if (s >= 70) { "C" } else { "F" } }; g(85)`, "B"},
		{`let g = fn(s) { if (s >= 90) { "A" } elseif (s >= 80) { "B" } else if (s >= 70) { "C" } else { "F" } }; g(72)`, "C"},
		{`let g = fn(s) { if (s >= 90) { "A" } elseif (s >= 80) { "B" } else if (s >= 70) { "C" } else { "F" } }; g(40)`, "F"},
	}

	runVMTests(t, tests)
}

func TestLoopCaptureVM(t *testing.T) {
	tests := []vmTestCase{
		{"let fns = []; for (i := 0; i < 3; i++) { fns = push(fns, fn() { i }) }; fns[0]() + fns[1]() + fns[2]()", 3},
		{"let fns = []; for (i := 0; i < 3; i++) { let t = i * 10; fns = push(fns, fn() { t }) }; fns[0]() + fns[1]() + fns[2]()", 30},
		{"let f = fn() { let fns = []; for (i := 0; i < 3; i++) { fns = push(fns, fn() { i }) }; fns[0]() + fns[1]() + fns[2]() }; f()", 3},
		{"let x = 0; for (i := 0; i < 3; i++) { x = x + i }; x;", 3},
	}

	runVMTests(t, tests)
}
