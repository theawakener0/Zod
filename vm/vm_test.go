package vm

import (
	"github.com/theawakener0/Zod/ast"
	lx "github.com/theawakener0/Zod/lexer"
	ps "github.com/theawakener0/Zod/parser"
	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/compiler"
	"fmt"
	"testing"
)

type vmTestCase struct {
	input 		string
	expected 	any
}

func TestIntegerArthimetic(t *testing.T) {
	tests := []vmTestCase {
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
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

		vm := New(comp.Bytecode())
		err1 := vm.Run()
		if err1 != nil {
			t.Fatalf("vm error: %s", err1)
		}

		stackElem := vm.StackTop()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(
	t 			*testing.T,
	expected 	any,
	actual 		obj.Object,
) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	}
}

func parse(input string)  *ast.Program {
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
