package vm

import (
	"testing"

	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

// runVMProgram compiles and runs input, requiring the run to succeed, and
// returns the last value popped from the stack.
func runVMProgram(t *testing.T, input string) obj.Object {
	t.Helper()

	program := parse(input)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error for %q: %s", input, err)
	}

	machine := New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		t.Fatalf("vm error for %q: %s", input, err)
	}

	return machine.LastPoppedStackElem()
}

// runVMFloatTests checks success cases whose expected value is a float64.
func runVMFloatTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		expected, ok := tt.expected.(float64)
		if !ok {
			t.Fatalf("expected float64 for %q, got %T", tt.input, tt.expected)
		}

		actual := runVMProgram(t, tt.input)
		result, ok := actual.(*obj.Float)
		if !ok {
			t.Errorf("object is not Float for %q. got=%T (%+v)", tt.input, actual, actual)
			continue
		}
		if result.Value != expected {
			t.Errorf("object has wrong value for %q. got=%v, want=%v", tt.input, result.Value, expected)
		}
	}
}

// runVMInspectTests checks the Inspect() representation of a success case.
func runVMInspectTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		expected, ok := tt.expected.(string)
		if !ok {
			t.Fatalf("expected string for %q, got %T", tt.input, tt.expected)
		}

		actual := runVMProgram(t, tt.input)
		if actual == nil {
			t.Errorf("no value for %q, want Inspect()=%q", tt.input, expected)
			continue
		}
		if actual.Inspect() != expected {
			t.Errorf("wrong Inspect() for %q. got=%q, want=%q", tt.input, actual.Inspect(), expected)
		}
	}
}

// runVMErrorTests accepts either a compile error or a VM abort and checks
// that the error message matches the expected message.
func runVMErrorTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		expected, ok := tt.expected.(string)
		if !ok {
			t.Fatalf("expected string for %q, got %T", tt.input, tt.expected)
		}

		program := parse(tt.input)
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			if err.Error() != expected {
				t.Errorf("wrong compiler error for %q. got=%q, want=%q", tt.input, err.Error(), expected)
			}
			continue
		}

		machine := New(comp.Bytecode())
		err := machine.Run()
		if err == nil {
			t.Errorf("expected compile error or vm abort for %q, got none", tt.input)
			continue
		}
		if err.Error() != expected {
			t.Errorf("wrong vm abort for %q. got=%q, want=%q", tt.input, err.Error(), expected)
		}
	}
}

func TestFloatExpressionVM(t *testing.T) {
	tests := []vmTestCase{
		{"3.14", 3.14},
		{"20.0", 20.0},
		{"20.", 20.0},
		{".5", 0.5},
		{"-3.14", -3.14},
		{"1.5 + 2.5", 4.0},
		{"5.0 - 1.5", 3.5},
		{"2.5 * 4", 10.0},
		{"10 / 4.0", 2.5},
		{"5 + 0.5", 5.5},
		{"5 * 1.5", 7.5},
		{"(1.5 + 2.5) * 2", 8.0},
		{"float(5) / 2", 2.5},
	}

	runVMFloatTests(t, tests)
}

func TestFloatComparisonVM(t *testing.T) {
	tests := []vmTestCase{
		{"1.5 < 2.5", true},
		{"2.5 < 1.5", false},
		{"2.5 > 1.5", true},
		{"1.5 > 2.5", false},
		{"1.5 == 1.5", true},
		{"1.5 == 1.6", false},
		{"1.5 != 1.6", true},
		{"1.5 != 1.5", false},
		{"1.5 <= 1.5", true},
		{"1.5 >= 1.5", true},
		{"1 == 1.0", true},
		{"1 != 1.0", false},
		{"1 < 1.5", true},
		{"2 > 1.5", true},
	}

	runVMTests(t, tests)
}

func TestFloatIncDecVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 1.5; ++x; x;", 2.5},
		{"let x = 1.5; --x; x;", 0.5},
		{"let x = 1.5; x++; x;", 2.5},
		{"let x = 1.5; x--; x;", 0.5},
		{"let x = 1.5; let y = x++; y;", 1.5},
		{"let x = 1.5; let y = ++x; y;", 2.5},
	}

	runVMFloatTests(t, tests)
}

func TestFloatConversionsVM(t *testing.T) {
	runVMTests(t, []vmTestCase{
		{"int(3.9)", 3},
		{"int(20.0)", 20},
		{"int(0.5)", 0},
		{"int(float(5))", 5},
	})

	runVMFloatTests(t, []vmTestCase{
		{"float(5)", 5.0},
		{"float(true)", 1.0},
		{"float(false)", 0.0},
		{`float("2.5")`, 2.5},
		{"float(2.5)", 2.5},
	})
}

func TestFloatInspectVM(t *testing.T) {
	tests := []vmTestCase{
		{"20.0", "20.0"},
		{"3.5", "3.5"},
		{".5", "0.5"},
	}

	runVMInspectTests(t, tests)
}

func TestHashBuiltinsVM(t *testing.T) {
	tests := []vmTestCase{
		{`len({"a": 1, "b": 2})`, 2},
		{`len({})`, 0},
		{`insert({"a": 1}, "b", 2)["b"]`, 2},
		{`len(insert({"a": 1}, "b", 2))`, 2},
		{`insert({"a": 1}, "a", 9)["a"]`, 9},
		{`let h = {"a": 1}; insert(h, "b", 2); len(h)`, 1},
		{`len(remove({"a": 1, "b": 2}, "a"))`, 1},
		{`remove({"a": 1, "b": 2}, "a")["b"]`, 2},
		{`remove({"a": 1, "b": 2}, "c")["a"]`, 1},
		{`contains({"a": 1}, "a")`, true},
		{`contains({"a": 1}, "b")`, false},
		{`contains({}, "a")`, false},
		{`contains(insert({}, "a", 1), "a")`, true},
		{`len(keys({"a": 1, "b": 2}))`, 2},
		{`len(vals({"a": 1, "b": 2}))`, 2},
		{`{"a": 1, "b": 2}["a"]`, 1},
		{`{"a": 1, "b": 2}["c"]`, Null},
		{`let h = {"a": 1}; h["b"] = 2; len(h)`, 2},
		{`let h = {"a": 1}; h["b"] = 2; h["b"]`, 2},
		{`let h = {"a": 1}; h["a"] = 9; h["a"]`, 9},
		{`let h = {"a": 1}; h["a"] += 5; h["a"]`, 6},
		{`let h = {5: 1}; h[5] = 2; h[5]`, 2},
	}

	runVMTests(t, tests)
}

func TestHashBuiltinErrorsVM(t *testing.T) {
	tests := []vmTestCase{
		{`insert({"a": 1})`, "wrong number of arguments. got=1, want=3"},
		{`insert([1, 2], "a", 1)`, "argument to `insert` not supported. got=ARRAY"},
		// VM functions are CLOSURE_OBJ values, so the hash-key error names
		// CLOSURE.
		{`insert({}, fn(x) { x }, 1)`, "unusable as hash key: CLOSURE"},
		{`remove({"a": 1})`, "wrong number of arguments. got=1, want=2"},
		{`remove(5, "a")`, "argument to `remove` not supported. got=INTEGER"},
		{`keys(1)`, "argument to `keys` not supported. got=INTEGER"},
		{`vals("str")`, "argument to `vals` not supported. got=STRING"},
		{`contains(1, 1)`, "argument to `contains` not supported. got=INTEGER"},
		{`let h = {}; h["a"] += 1`, "key not found: a"},
	}

	runVMErrorTests(t, tests)
}

func TestHashPreservesOrderVM(t *testing.T) {
	tests := []vmTestCase{
		{`{"b": 2, "a": 1, "c": 3}`, "{b: 2, a: 1, c: 3}"},
	}

	runVMInspectTests(t, tests)
}

func TestHashKeysValsOrderVM(t *testing.T) {
	tests := []vmTestCase{
		{`keys({"b": 2, "a": 1, "c": 3})`, "[b, a, c]"},
		{`vals({"b": 2, "a": 1, "c": 3})`, "[2, 1, 3]"},
	}

	runVMInspectTests(t, tests)
}

func TestHashInsertRemoveOrderVM(t *testing.T) {
	tests := []vmTestCase{
		{`insert({"b": 2, "a": 1}, "c", 3)`, "{b: 2, a: 1, c: 3}"},
		{`remove({"b": 2, "a": 1, "c": 3}, "a")`, "{b: 2, c: 3}"},
	}

	runVMInspectTests(t, tests)
}

func TestHashIndexAssignmentOrderVM(t *testing.T) {
	input := `
	let h = {"x": 1}
	h["y"] = 2
	h["z"] = 3
	h`
	tests := []vmTestCase{
		{input, "{x: 1, y: 2, z: 3}"},
	}

	runVMInspectTests(t, tests)
}

func TestNestedLoopVariableIsolationVM(t *testing.T) {
	input := `
	let seen = ""
	for (i := 0; i < 3; i++) {
		for (i := 0; i < 3; i++) {
			seen = seen + "i"
		}
	}
	seen`
	tests := []vmTestCase{
		{input, "iiiiiiiii"},
	}

	runVMTests(t, tests)
}

func TestLoopVariableDoesNotLeakVM(t *testing.T) {
	tests := []vmTestCase{
		{
			`let x = 0; for (i := 0; i < 3; i++) { x = x + 1 }; if (i == 3) { x = 100 }; x`,
			"identifier not found: i",
		},
	}

	runVMErrorTests(t, tests)
}

func TestBreakContinueOutsideLoopVM(t *testing.T) {
	tests := []vmTestCase{
		{"break;", "break used outside of loop"},
		{"continue;", "continue used outside of loop"},
		{"let f = fn() { break; }; f();", "break used outside of loop"},
	}

	runVMErrorTests(t, tests)
}

func TestForExpressionVM(t *testing.T) {
	tests := []vmTestCase{
		{"let f = fn() { let i = 0; for (; i < 5; ++i) { if (i == 3) { return i; } } }; f()", 3},
		{"let x = 0; for (let i = 0; i < 5; i++) { x = x + i; }; x;", 10},
		{"let x = 0; for (x = 0; x < 5; x += 1) { }; x;", 5},
	}

	runVMTests(t, tests)
}

func TestForWhileExpressionVM(t *testing.T) {
	tests := []vmTestCase{
		{"let f = fn() { let i = 0; for (i < 3) { if (i == 2) { return i; }; ++i; } }; f()", 2},
	}

	runVMTests(t, tests)
}

func TestLoopExpressionVM(t *testing.T) {
	tests := []vmTestCase{
		{"let f = fn() { let i = 0; loop { if (i == 5) { return i; }; ++i; } }; f()", 5},
	}

	runVMTests(t, tests)
}

func TestPrefixIncDecVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 5; ++x; x;", 6},
		{"let x = 5; --x; x;", 4},
		{"let x = 5; ++x; ++x; x;", 7},
		{"let x = 5; --x; --x; x;", 3},
	}

	runVMTests(t, tests)
}

func TestPostfixIncDecVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 5; x++; x;", 6},
		{"let x = 5; x--; x;", 4},
		{"let x = 5; x++; x++; x;", 7},
	}

	runVMTests(t, tests)
}

func TestPostfixIncDecValueVM(t *testing.T) {
	tests := []vmTestCase{
		{"let x = 5; let y = x++; y;", 5},
		{"let x = 5; let y = x--; y;", 5},
		{"let x = 5; let y = x++; y; x;", 6},
		{"let x = 5; let y = x--; y; x;", 4},
		{"let x = 5; let y = ++x; y;", 6},
		{"let x = 5; let y = --x; y;", 4},
	}

	runVMTests(t, tests)
}

func TestMatrixCreationVM(t *testing.T) {
	actual := runVMProgram(t, `matrix(2, 3, [1, 2, 3, 4, 5, 6])`)
	m, ok := actual.(*obj.Matrix)
	if !ok {
		t.Fatalf("matrix did not return Matrix. got=%T (%+v)", actual, actual)
	}
	if m.Rows != 2 || m.Cols != 3 {
		t.Fatalf("matrix dimensions wrong. got=%dx%d, want=2x3", m.Rows, m.Cols)
	}
	if m.Inspect() != "[[1, 2, 3], [4, 5, 6]]" {
		t.Errorf("matrix Inspect wrong. got=%s", m.Inspect())
	}
}

func TestMatrixFloatDimensionsVM(t *testing.T) {
	actual := runVMProgram(t, `matrix(2.0, 2.0, [1, 2, 3, 4])`)
	m, ok := actual.(*obj.Matrix)
	if !ok {
		t.Fatalf("matrix did not return Matrix. got=%T (%+v)", actual, actual)
	}
	if m.Rows != 2 || m.Cols != 2 {
		t.Fatalf("matrix dimensions wrong. got=%dx%d, want=2x2", m.Rows, m.Cols)
	}
}

func TestMatrixErrorsVM(t *testing.T) {
	tests := []vmTestCase{
		{`matrix(2, 3, [1, 2, 3])`, "array length does not match matrix dimensions. got=3, want=6"},
		{`matrix(0, 3, [])`, "matrix dimensions must be positive. got=0, 3"},
		{`matrix(2, 2, ["a", "b", "c", "d"])`, "matrix elements must be integers or floats. got=STRING"},
		{`matrix(2, 2, [1, 2, 3, 4]) + matrix(3, 3, [1, 1, 1, 1, 1, 1, 1, 1, 1])`, "matrix dimension mismatch for +: 2x2 vs 3x3"},
		{`matrix(2, 2, [1, 2, 3, 4]) * matrix(3, 3, [1, 1, 1, 1, 1, 1, 1, 1, 1])`, "matrix dimension mismatch for multiplication: 2x2 vs 3x3"},
	}

	runVMErrorTests(t, tests)
}

func TestTryProgramDoesNotAbortVM(t *testing.T) {
	input := `
	let out = ""
	let r = try(int("abc"))
	if (r[0]) {
		out = "parsed"
	} else {
		out = "failed: " + r[1]
	}
	out`
	tests := []vmTestCase{
		{input, `failed: could not parse "abc" as integer`},
	}

	runVMTests(t, tests)
}

func TestTryWrongArgCountVM(t *testing.T) {
	// The VM compiler deliberately turns a bad try() arity into a caught
	// error pair, so the program continues and observes [false, message].
	actual := runVMProgram(t, `try()`)
	arr, ok := actual.(*obj.Array)
	if !ok {
		t.Fatalf("try() did not return Array. got=%T (%+v)", actual, actual)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("try() array has wrong length. got=%d, want=2", len(arr.Elements))
	}

	okFlag, ok := arr.Elements[0].(*obj.Boolean)
	if !ok {
		t.Fatalf("try() ok flag is not Boolean. got=%T", arr.Elements[0])
	}
	if okFlag.Value {
		t.Errorf("try() ok flag wrong. got=true, want=false")
	}

	message, ok := arr.Elements[1].(*obj.String)
	if !ok {
		t.Fatalf("try() error message is not String. got=%T", arr.Elements[1])
	}
	if message.Value != "wrong number of arguments. got=0, want=1" {
		t.Errorf("try() error message wrong. got=%q", message.Value)
	}
}

func TestNilReturningStatementsVM(t *testing.T) {
	tests := []vmTestCase{
		{`let f = fn() { let x = 1; }; f();`, Null},
		{`let f = fn() { a := 2; }; f();`, Null},
		{`let f = fn() { let y = [1, 2]; y[0] = 5; }; f();`, Null},
		{`let f = fn() { let h = {"a": 1}; h["a"] = 2; }; f();`, Null},
	}

	runVMTests(t, tests)
}
