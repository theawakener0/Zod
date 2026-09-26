package main

import (
	"testing"

	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/compiler"
	eval "github.com/theawakener0/Zod/evaluator"
	lx "github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	ps "github.com/theawakener0/Zod/parser"
	"github.com/theawakener0/Zod/vm"
)

// Representative programs. Kept small so each op stays well under 5s.
const (
	benchFib15Src = "fibonacci := fn(x) { if (x==0) { return 0; } else { if (x==1) { return 1; } else { fibonacci(x-1) + fibonacci(x-2) } } }; fibonacci(15)"

	benchArithLoopSrc = "let x = 0; for (i := 0; i < 200; i++) { x = x + i; }; x;"

	benchArrayPushSrc = "let a = []; for (i := 0; i < 100; i++) { a = push(a, i); }; len(a);"

	benchHashInsertSrc = "let h = {}; for (i := 0; i < 100; i++) { h = insert(h, i, i * 2); }; len(h);"

	benchStringConcatSrc = `let s = ""; for (i := 0; i < 100; i++) { s = s + "x"; }; len(s);`

	benchMatrixCellSrc = "let g = matrix(48, 120, make(48*120, 1)); let s = 0; let rows = len(g); let cols = len(g[0]); for (r := 0; r < rows; r++) { for (c := 0; c < cols; c++) { s = s + g[r][c]; } }; s;"
)

// benchSink prevents the compiler from eliminating benchmarked work.
var benchSink obj.Object

func benchParse(b *testing.B, src string) (*ast.Program, bool) {
	b.Helper()
	l := lx.New(src)
	p := ps.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return nil, false
	}
	return program, true
}

func benchCompile(b *testing.B, src string) (*compiler.Bytecode, bool) {
	b.Helper()
	program, ok := benchParse(b, src)
	if !ok {
		return nil, false
	}
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		return nil, false
	}
	return comp.Bytecode(), true
}

func benchRunVM(bytecode *compiler.Bytecode) (obj.Object, error) {
	machine := vm.New(bytecode)
	if err := machine.Run(); err != nil {
		return nil, err
	}
	return machine.LastPoppedStackElem(), nil
}

func benchRunEval(program *ast.Program) obj.Object {
	env := obj.NewEnviroment()
	return eval.Eval(program, env)
}

func benchIsError(o obj.Object) bool {
	if o == nil {
		return true
	}
	_, ok := o.(*obj.Error)
	return ok
}

func benchmarkVM(b *testing.B, src string, fatal bool) {
	b.Helper()
	b.ReportAllocs()
	bytecode, ok := benchCompile(b, src)
	if !ok {
		if fatal {
			b.Fatalf("vm compile failed for %q", src)
		}
		b.Skipf("vm compile unsupported, skipping: %q", src)
	}
	if _, err := benchRunVM(bytecode); err != nil {
		if fatal {
			b.Fatalf("vm trial run failed: %s", err)
		}
		b.Skipf("vm trial run unsupported, skipping: %s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine := vm.New(bytecode)
		if err := machine.Run(); err != nil {
			b.Fatalf("vm error: %s", err)
		}
		benchSink = machine.LastPoppedStackElem()
	}
}

func benchmarkEval(b *testing.B, src string, fatal bool) {
	b.Helper()
	b.ReportAllocs()
	program, ok := benchParse(b, src)
	if !ok {
		if fatal {
			b.Fatalf("eval parse failed for %q", src)
		}
		b.Skipf("eval parse unsupported, skipping: %q", src)
	}
	if got := benchRunEval(program); benchIsError(got) {
		if fatal {
			b.Fatalf("eval trial run returned error: %s", got.Inspect())
		}
		b.Skipf("eval trial run unsupported, skipping: %s", got.Inspect())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := obj.NewEnviroment()
		got := eval.Eval(program, env)
		if errObj, isErr := got.(*obj.Error); isErr {
			b.Fatalf("eval error: %s", errObj.Message)
		}
		benchSink = got
	}
}

// BenchmarkFibVM15 compiles fib(15) once, then measures VM execution only.
func BenchmarkFibVM15(b *testing.B) {
	benchmarkVM(b, benchFib15Src, true)
}

// BenchmarkFibEval15 parses fib(15) once, then measures tree-walking eval only.
func BenchmarkFibEval15(b *testing.B) {
	benchmarkEval(b, benchFib15Src, true)
}

// BenchmarkArithLoop covers both engines via sub-benchmarks.
func BenchmarkArithLoop(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchArithLoopSrc, true)
	})
	b.Run("eval", func(b *testing.B) {
		benchmarkEval(b, benchArithLoopSrc, true)
	})
}

// BenchmarkArithLoopVM measures a simple integer for-loop on the VM.
func BenchmarkArithLoopVM(b *testing.B) {
	benchmarkVM(b, benchArithLoopSrc, true)
}

// BenchmarkArithLoopEval measures a simple integer for-loop on the evaluator.
func BenchmarkArithLoopEval(b *testing.B) {
	benchmarkEval(b, benchArithLoopSrc, true)
}

// BenchmarkArrayPush covers both engines via sub-benchmarks; skips gracefully
// if the VM does not support the push builtin.
func BenchmarkArrayPush(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchArrayPushSrc, false)
	})
	b.Run("eval", func(b *testing.B) {
		benchmarkEval(b, benchArrayPushSrc, false)
	})
}

// BenchmarkArrayPushVM measures repeated push() in a loop on the VM.
func BenchmarkArrayPushVM(b *testing.B) {
	benchmarkVM(b, benchArrayPushSrc, false)
}

// BenchmarkArrayPushEval measures repeated push() in a loop on the evaluator.
func BenchmarkArrayPushEval(b *testing.B) {
	benchmarkEval(b, benchArrayPushSrc, false)
}

// BenchmarkHashInsert covers both engines via sub-benchmarks; skips gracefully
// if the VM does not support the insert builtin.
func BenchmarkHashInsert(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchHashInsertSrc, false)
	})
	b.Run("eval", func(b *testing.B) {
		benchmarkEval(b, benchHashInsertSrc, false)
	})
}

// BenchmarkHashInsertVM measures repeated insert() in a loop on the VM.
func BenchmarkHashInsertVM(b *testing.B) {
	benchmarkVM(b, benchHashInsertSrc, false)
}

// BenchmarkHashInsertEval measures repeated insert() in a loop on the evaluator.
func BenchmarkHashInsertEval(b *testing.B) {
	benchmarkEval(b, benchHashInsertSrc, false)
}

// BenchmarkStringConcat covers both engines via sub-benchmarks.
func BenchmarkStringConcat(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchStringConcatSrc, false)
	})
	b.Run("eval", func(b *testing.B) {
		benchmarkEval(b, benchStringConcatSrc, false)
	})
}

// BenchmarkStringConcatVM measures repeated string "+" in a loop on the VM.
func BenchmarkStringConcatVM(b *testing.B) {
	benchmarkVM(b, benchStringConcatSrc, false)
}

// BenchmarkStringConcatEval measures repeated string "+" in a loop on the evaluator.
func BenchmarkStringConcatEval(b *testing.B) {
	benchmarkEval(b, benchStringConcatSrc, false)
}

// BenchmarkMatrixCell covers both engines via sub-benchmarks; skips gracefully
// if the VM does not support builtins used by the canvas loop.
func BenchmarkMatrixCell(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchMatrixCellSrc, false)
	})
	b.Run("eval", func(b *testing.B) {
		benchmarkEval(b, benchMatrixCellSrc, false)
	})
}

// BenchmarkMatrixCellVM measures a 48x120 matrix double-index read loop on the VM.
func BenchmarkMatrixCellVM(b *testing.B) {
	benchmarkVM(b, benchMatrixCellSrc, false)
}

// BenchmarkMatrixCellEval measures a 48x120 matrix double-index read loop on the evaluator.
func BenchmarkMatrixCellEval(b *testing.B) {
	benchmarkEval(b, benchMatrixCellSrc, false)
}
