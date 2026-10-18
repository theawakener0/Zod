package repl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/vm"
)

var testEngines = []string{"--eng=vm", "--eng=eval"}

func TestExecuteSuccessBothEngines(t *testing.T) {
	for _, engine := range testEngines {
		var out bytes.Buffer
		err := Execute("1 + 2", &out, engine, ".")
		if err != nil {
			t.Fatalf("engine %s: Execute(1 + 2) err = %v, want nil", engine, err)
		}
		if !strings.Contains(out.String(), "3") {
			t.Fatalf("engine %s: Execute(1 + 2) out = %q, want to contain %q", engine, out.String(), "3")
		}
	}
}

func TestExecuteParseFailureBothEngines(t *testing.T) {
	for _, engine := range testEngines {
		var out bytes.Buffer
		err := Execute("let x = \n", &out, engine, ".")
		if err == nil {
			t.Fatalf("engine %s: Execute(parse failure) err = nil, want non-nil", engine)
		}
		if !strings.Contains(err.Error(), "Parse errors") {
			t.Fatalf("engine %s: Execute(parse failure) err = %q, want to contain %q", engine, err.Error(), "Parse errors")
		}
		if out.Len() != 0 {
			t.Fatalf("engine %s: Execute(parse failure) out = %q, want empty (no diagnostics on stdout)", engine, out.String())
		}
	}
}

func TestExecuteRuntimeErrorBothEngines(t *testing.T) {
	for _, engine := range testEngines {
		var out bytes.Buffer
		err := Execute(`error("boom")`, &out, engine, ".")
		if err == nil {
			t.Fatalf("engine %s: Execute(error boom) err = nil, want non-nil", engine)
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("engine %s: Execute(error boom) err = %q, want to contain %q", engine, err.Error(), "boom")
		}
		if out.Len() != 0 {
			t.Fatalf("engine %s: Execute(error boom) out = %q, want empty", engine, out.String())
		}
	}
}

func TestExecuteTrySwallowsBothEngines(t *testing.T) {
	for _, engine := range testEngines {
		var out bytes.Buffer
		err := Execute(`try(error("x"))`, &out, engine, ".")
		if err != nil {
			t.Fatalf("engine %s: Execute(try(error)) err = %v, want nil", engine, err)
		}
	}
}

func TestExecuteUnknownEngine(t *testing.T) {
	var out bytes.Buffer
	err := Execute("1 + 2", &out, "--eng=bogus", ".")
	if err == nil {
		t.Fatalf("Execute(unknown engine) err = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "unknown engine") {
		t.Fatalf("Execute(unknown engine) err = %q, want to contain %q", err.Error(), "unknown engine")
	}
}

func TestRunVMStateSurvivesErrors(t *testing.T) {
	constants := make([]obj.Object, 0, 256)
	globals := make([]obj.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	var out bytes.Buffer
	updated, err := runVM("42", &out, constants, globals, symbolTable, ".")
	if err != nil {
		t.Fatalf("runVM(42) err = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("runVM(42) out = %q, want to contain %q", out.String(), "42")
	}
	if updated == nil {
		t.Fatalf("runVM(42) returned nil constants, want usable slice")
	}

	out.Reset()
	afterParseErr, err := runVM("let x = \n", &out, updated, globals, symbolTable, ".")
	if err == nil {
		t.Fatalf("runVM(parse failure) err = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "Parse errors") {
		t.Fatalf("runVM(parse failure) err = %q, want to contain %q", err.Error(), "Parse errors")
	}
	if out.Len() != 0 {
		t.Fatalf("runVM(parse failure) out = %q, want empty", out.String())
	}
	if afterParseErr == nil {
		t.Fatalf("runVM(parse failure) returned nil constants, want usable slice")
	}

	out.Reset()
	afterRuntimeErr, err := runVM(`error("boom")`, &out, afterParseErr, globals, symbolTable, ".")
	if err == nil {
		t.Fatalf("runVM(runtime error) err = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("runVM(runtime error) err = %q, want to contain %q", err.Error(), "boom")
	}
	if out.Len() != 0 {
		t.Fatalf("runVM(runtime error) out = %q, want empty", out.String())
	}
	if afterRuntimeErr == nil {
		t.Fatalf("runVM(runtime error) returned nil constants, want usable slice")
	}

	out.Reset()
	_, err = runVM("42", &out, afterRuntimeErr, globals, symbolTable, ".")
	if err != nil {
		t.Fatalf("runVM(42) after errors err = %v, want nil (state must survive errors)", err)
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("runVM(42) after errors out = %q, want to contain %q", out.String(), "42")
	}
}

func TestRunEvalStateSurvivesErrors(t *testing.T) {
	env := obj.NewEnviroment()

	var out bytes.Buffer
	if err := runEval("42", env, &out); err != nil {
		t.Fatalf("runEval(42) err = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("runEval(42) out = %q, want to contain %q", out.String(), "42")
	}

	out.Reset()
	if err := runEval("let x = \n", env, &out); err == nil {
		t.Fatalf("runEval(parse failure) err = nil, want non-nil")
	} else if !strings.Contains(err.Error(), "Parse errors") {
		t.Fatalf("runEval(parse failure) err = %q, want to contain %q", err.Error(), "Parse errors")
	}
	if out.Len() != 0 {
		t.Fatalf("runEval(parse failure) out = %q, want empty", out.String())
	}

	out.Reset()
	if err := runEval(`error("boom")`, env, &out); err == nil {
		t.Fatalf("runEval(runtime error) err = nil, want non-nil")
	} else if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("runEval(runtime error) err = %q, want to contain %q", err.Error(), "boom")
	}
	if out.Len() != 0 {
		t.Fatalf("runEval(runtime error) out = %q, want empty", out.String())
	}

	out.Reset()
	if err := runEval("42", env, &out); err != nil {
		t.Fatalf("runEval(42) after errors err = %v, want nil (env must survive errors)", err)
	}
	if !strings.Contains(out.String(), "42") {
		t.Fatalf("runEval(42) after errors out = %q, want to contain %q", out.String(), "42")
	}
}

func TestForwardReferencesBothEngines(t *testing.T) {
	const src = `let is_even = fn(n) { if (n == 0) { return true; } else { return is_odd(n - 1); } };
let is_odd = fn(n) { if (n == 0) { return false; } else { return is_even(n - 1); } };
is_even(10);`

	for _, engine := range testEngines {
		var out bytes.Buffer
		if err := Execute(src, &out, engine, "."); err != nil {
			t.Fatalf("engine %s: Execute(mutual recursion) err = %v, want nil", engine, err)
		}
		if !strings.Contains(out.String(), "true") {
			t.Fatalf("engine %s: out = %q, want to contain %q", engine, out.String(), "true")
		}
	}
}

func TestNestedIndexAssignBothEngines(t *testing.T) {
	const src = `let h = {"a": {"b": 0}}; h["a"]["b"] = 5; h["a"]["b"];`

	for _, engine := range testEngines {
		var out bytes.Buffer
		if err := Execute(src, &out, engine, "."); err != nil {
			t.Fatalf("engine %s: Execute(nested index assign) err = %v, want nil", engine, err)
		}
		if !strings.Contains(out.String(), "5") {
			t.Fatalf("engine %s: out = %q, want to contain %q", engine, out.String(), "5")
		}
	}
}

func TestNilReturningStatementsBothEngines(t *testing.T) {
	const src = `let f = fn() { let x = 1; }; f();`

	for _, engine := range testEngines {
		var out bytes.Buffer
		err := Execute(src, &out, engine, ".")
		if err != nil {
			t.Fatalf("engine %s: Execute(let-only body) err = %v, want nil", engine, err)
		}
		if out.Len() != 0 {
			t.Fatalf("engine %s: out = %q, want empty (null results are not printed)", engine, out.String())
		}
	}
}
