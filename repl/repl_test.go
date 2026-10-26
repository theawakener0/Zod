package repl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/vm"
)

func TestExecuteSuccess(t *testing.T) {
	var out bytes.Buffer
	err := Execute("1 + 2", &out, ".")
	if err != nil {
		t.Fatalf("Execute(1 + 2) err = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "3") {
		t.Fatalf("Execute(1 + 2) out = %q, want to contain %q", out.String(), "3")
	}
}

func TestExecuteParseFailure(t *testing.T) {
	var out bytes.Buffer
	err := Execute("let x = \n", &out, ".")
	if err == nil {
		t.Fatalf("Execute(parse failure) err = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "Parse errors") {
		t.Fatalf("Execute(parse failure) err = %q, want to contain %q", err.Error(), "Parse errors")
	}
	if out.Len() != 0 {
		t.Fatalf("Execute(parse failure) out = %q, want empty (no diagnostics on stdout)", out.String())
	}
}

func TestExecuteRuntimeError(t *testing.T) {
	var out bytes.Buffer
	err := Execute(`error("boom")`, &out, ".")
	if err == nil {
		t.Fatalf("Execute(error boom) err = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("Execute(error boom) err = %q, want to contain %q", err.Error(), "boom")
	}
	if out.Len() != 0 {
		t.Fatalf("Execute(error boom) out = %q, want empty", out.String())
	}
}

func TestExecuteTrySwallows(t *testing.T) {
	var out bytes.Buffer
	err := Execute(`try(error("x"))`, &out, ".")
	if err != nil {
		t.Fatalf("Execute(try(error)) err = %v, want nil", err)
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

func TestForwardReferences(t *testing.T) {
	const src = `let is_even = fn(n) { if (n == 0) { return true; } else { return is_odd(n - 1); } };
let is_odd = fn(n) { if (n == 0) { return false; } else { return is_even(n - 1); } };
is_even(10);`

	var out bytes.Buffer
	if err := Execute(src, &out, "."); err != nil {
		t.Fatalf("Execute(mutual recursion) err = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "true") {
		t.Fatalf("out = %q, want to contain %q", out.String(), "true")
	}
}

func TestNestedIndexAssign(t *testing.T) {
	const src = `let h = {"a": {"b": 0}}; h["a"]["b"] = 5; h["a"]["b"];`

	var out bytes.Buffer
	if err := Execute(src, &out, "."); err != nil {
		t.Fatalf("Execute(nested index assign) err = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "5") {
		t.Fatalf("out = %q, want to contain %q", out.String(), "5")
	}
}

func TestNilReturningStatements(t *testing.T) {
	const src = `let f = fn() { let x = 1; }; f();`

	var out bytes.Buffer
	err := Execute(src, &out, ".")
	if err != nil {
		t.Fatalf("Execute(let-only body) err = %v, want nil", err)
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q, want empty (null results are not printed)", out.String())
	}
}
