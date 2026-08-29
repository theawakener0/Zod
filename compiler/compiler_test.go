package compiler

import (
	"github.com/theawakener0/Zod/ast"
	lx "github.com/theawakener0/Zod/lexer"
	ps "github.com/theawakener0/Zod/parser"
	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/code"
	"fmt"
	"testing"
)

type compilerTestCase struct {
	input 					string
	expectedConstants		[]any
	expectedInstructions 	[]code.Instructions
}

func TestIntergerArthimetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		
		compiler := New()
		err0 := compiler.Compile(program)
		if err0 != nil {
			t.Fatalf("compile error: %s", err0)
		}

		bytecode := compiler.Bytecode()

		err1 := testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err1 != nil {
			t.Fatalf("testInstructions failed: %s", err1)
		}

		err2 := testConstant(tt.expectedConstants, bytecode.Constant)
		if err2 != nil {
			t.Fatalf("testConstant failed: %s", err2)
		}
	}
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

func testConstant(
	expected []any,
	actual []obj.Object,
) error {
	if len(actual) != len(expected) {
		return  fmt.Errorf("wrong number of constants. got=%d, want=%d", actual, expected)
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		}
	}

	return nil
}

func testInstructions(
	expected []code.Instructions,
	actual code.Instructions,
) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instrucions lenght.\nwant=%q\ngot=%q", concatted, actual)
	}

	for i, instruction := range concatted {
		if actual[i] != instruction {
			return  fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, instruction, actual[i])
		}
	}

	return nil
}

func concatInstructions(instrucions []code.Instructions)  code.Instructions {
	out := code.Instructions{}

	for _, instruction := range instrucions {
		out = append(out, instruction...) 
	}

	return out
}

func parse(input string)  *ast.Program {
	l := lx.New(input)
	p := ps.New(l)
	return p.ParseProgram()
}
