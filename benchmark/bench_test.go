package main

import (
	"testing"

	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/compiler"
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

// BenchmarkFibVM15 compiles fib(15) once, then measures VM execution only.
func BenchmarkFibVM15(b *testing.B) {
	benchmarkVM(b, benchFib15Src, true)
}

// BenchmarkArithLoop measures a simple integer for-loop on the VM.
func BenchmarkArithLoop(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchArithLoopSrc, true)
	})
}

// BenchmarkArithLoopVM measures a simple integer for-loop on the VM.
func BenchmarkArithLoopVM(b *testing.B) {
	benchmarkVM(b, benchArithLoopSrc, true)
}

// BenchmarkArrayPush measures repeated push() in a loop on the VM; skips
// gracefully if the VM does not support the push builtin.
func BenchmarkArrayPush(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchArrayPushSrc, false)
	})
}

// BenchmarkArrayPushVM measures repeated push() in a loop on the VM.
func BenchmarkArrayPushVM(b *testing.B) {
	benchmarkVM(b, benchArrayPushSrc, false)
}

// BenchmarkHashInsert measures repeated insert() in a loop on the VM; skips
// gracefully if the VM does not support the insert builtin.
func BenchmarkHashInsert(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchHashInsertSrc, false)
	})
}

// BenchmarkHashInsertVM measures repeated insert() in a loop on the VM.
func BenchmarkHashInsertVM(b *testing.B) {
	benchmarkVM(b, benchHashInsertSrc, false)
}

// BenchmarkStringConcat measures repeated string "+" in a loop on the VM.
func BenchmarkStringConcat(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchStringConcatSrc, false)
	})
}

// BenchmarkStringConcatVM measures repeated string "+" in a loop on the VM.
func BenchmarkStringConcatVM(b *testing.B) {
	benchmarkVM(b, benchStringConcatSrc, false)
}

// BenchmarkMatrixCell measures a 48x120 matrix double-index read loop on the
// VM; skips gracefully if the VM does not support builtins used by the canvas
// loop.
func BenchmarkMatrixCell(b *testing.B) {
	b.ReportAllocs()
	b.Run("vm", func(b *testing.B) {
		benchmarkVM(b, benchMatrixCellSrc, false)
	})
}

// BenchmarkMatrixCellVM measures a 48x120 matrix double-index read loop on the VM.
func BenchmarkMatrixCellVM(b *testing.B) {
	benchmarkVM(b, benchMatrixCellSrc, false)
}
