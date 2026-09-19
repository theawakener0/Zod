package compiler

import (
	"fmt"
	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/code"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code.Instructions
}

func TestIntergerArthimetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.0 + 2.0",
			expectedConstants: []any{1.0, 2.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.0; 2.0",
			expectedConstants: []any{1.0, 2.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.0 / 2",
			expectedConstants: []any{1.0, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) { 10 } else { 20 } 3333",
			expectedConstants: []any{10, 20, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpNotTruthy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 13),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "if (true) { 20 } 6666",
			expectedConstants: []any{20, 6666},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpNotTruthy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalAssigningStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "let one = 1; two := 2;",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input:             "let one = 1; one;",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "let one = 1; two := one; two;",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"hello"`,
			expectedConstants: []any{"hello"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"hello" + "world!"`,
			expectedConstants: []any{"hello", "world!"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []any{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 - 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpSub),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []any{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []any{1, 2, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunction(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { return 5 + 10 }",
			expectedConstants: []any{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { return 5 + 10 }",
			expectedConstants: []any{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { 1; 2 }",
			expectedConstants: []any{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { }",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { 24 }()",
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
		{
			input: "noArgs := fn() { 24 }; noArgs()",
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScope(t *testing.T) {
	comp := New()
	if comp.scopesIndex != 0 {
		t.Errorf("scopesIndex wrong. got=%d, want=%d", comp.scopesIndex, 0)
	}

	comp.emit(code.OpMul)

	comp.enterScope()
	if comp.scopesIndex != 1 {
		t.Errorf("scopesIndex wrong. got=%d, want=%d", comp.scopesIndex, 1)
	}

	comp.emit(code.OpSub)

	if len(comp.scopes[comp.scopesIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopesIndex].instructions))
	}

	last := comp.scopes[comp.scopesIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.OpCode wrong. got=%d, want=%d", last.Opcode, code.OpSub)
	}

	comp.leaveScope()
	if comp.scopesIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopesIndex, 0)
	}

	comp.emit(code.OpAdd)

	if len(comp.scopes[comp.scopesIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopesIndex].instructions))
	}

	last = comp.scopes[comp.scopesIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.OpCode wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}

	previous := comp.scopes[comp.scopesIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.OpCode wrong. got=%d, want=%d", previous.Opcode, code.OpMul)
	}
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "num := 55; fn() { num }",
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { let num = 55; num }",
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { let a = 55; b := 77; a + b}",
			expectedConstants: []any{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpDefineLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	comp := New()
	if comp.scopesIndex != 0 {
		t.Errorf("scopesIndex wrong. got=%d, want=%d", comp.scopesIndex, 0)
	}
	globalSymbolTable := comp.symbolTable

	comp.emit(code.OpMul)

	comp.enterScope()
	if comp.scopesIndex != 1 {
		t.Errorf("scopesIndex wrong. got=%d, want=%d", comp.scopesIndex, 1)
	}

	comp.emit(code.OpSub)

	if len(comp.scopes[comp.scopesIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopesIndex].instructions))
	}

	last := comp.scopes[comp.scopesIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.OpCode wrong. got=%d, want=%d", last.Opcode, code.OpSub)
	}

	if comp.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	comp.leaveScope()
	if comp.scopesIndex != 0 {
		t.Errorf("scopesIndex wrong. got=%d, want=%d", comp.scopesIndex, 0)
	}

	if comp.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if comp.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	comp.emit(code.OpAdd)

	if len(comp.scopes[comp.scopesIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopesIndex].instructions))
	}

	last = comp.scopes[comp.scopesIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.OpCode wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}

	previous := comp.scopes[comp.scopesIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.OpCode wrong. got=%d, want=%d", previous.Opcode, code.OpMul)
	}
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "oneArgs := fn(a) { }; oneArgs(24)",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "oneArgs := fn(a) { a }; oneArgs(24)",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let manyArgs = fn(a, b, c) { }; manyArgs(24, 25, 26)",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let manyArgs = fn(a, b, c) { a; b; c }; manyArgs(24, 25, 26)",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBuiltin(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "len([]); push([], 1);",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 11),
				code.Make(code.OpArray, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn() { len([]) }",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpArray, 0),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn(a) { fn(b) { fn(c) { a + b + c } } }",
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetFreeCell, 0),
					code.Make(code.OpGetLocalCell, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocalCell, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let global = 55;
			fn() {
				let a = 66;
				fn() {
					let b = 77;
					fn() {
						let c = 88;
						global + a + b + c;
					}
				}
			}
			`,
			expectedConstants: []any{
				55,
				66,
				77,
				88,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpGetFreeCell, 0),
					code.Make(code.OpGetLocalCell, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpGetLocalCell, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "countDown := fn(x) { countDown(x - 1) }; countDown(1)",
			expectedConstants: []any{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSub),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let wrapper = fn() {
				let countDown = fn(x) { countDown(x - 1); };
				countDown(1);
			};
			wrapper();
			`,
			expectedConstants: []any{
				1,
				[]code.Instructions{
					code.Make(code.OpCurrentClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSub),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
				[]code.Instructions{
					code.Make(code.OpClosure, 1, 0),
					code.Make(code.OpDefineLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 2),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 3, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestNull(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "null",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNull),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerIncrement(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `let a = 1; a++`,
			expectedConstants: []any{1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpDup),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
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

func testStringObject(expected string, actual obj.Object) error {
	result, ok := actual.(*obj.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
	}

	return nil
}

func testConstant(
	expected []any,
	actual []obj.Object,
) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d", actual, expected)
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(string(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*obj.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
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
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, instruction, actual[i])
		}
	}

	return nil
}

func concatInstructions(instrucions []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, instruction := range instrucions {
		out = append(out, instruction...)
	}

	return out
}

func parse(input string) *ast.Program {
	l := lx.New(input)
	p := ps.New(l)
	return p.ParseProgram()
}
