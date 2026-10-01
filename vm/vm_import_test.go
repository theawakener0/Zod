package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

// writeVMImportFixture creates a fresh temp module mirroring the shape of
// examples/pub_lib.zd: public `version` + `greet`, private `secret`.
// Every test gets its own t.TempDir() because the VM module caches
// (vmModuleCache/vmLoading in vm.go:25-26) are process-global and keyed by
// path, so paths must never be reused across tests.
func writeVMImportFixture(t *testing.T) (dir, modPath string) {
	t.Helper()
	dir = t.TempDir()
	modPath = filepath.Join(dir, "pub_lib.zd")
	content := "pub let version = 1\n" +
		"let secret = \"hidden\"\n" +
		"pub greet := fn(name) { \"hello, \" + name }\n"
	if err := os.WriteFile(modPath, []byte(content), 0644); err != nil {
		t.Fatalf("could not write temp module: %s", err)
	}
	return dir, modPath
}

// runVMWithBaseDir runs the full parse→compile→run pipeline on the VM with
// the compiler BaseDir set to dir. It returns the machine (nil when compile
// fails), the compile error, and the Run error.
func runVMWithBaseDir(t *testing.T, dir, input string) (*VM, error, error) {
	t.Helper()
	program := parse(input)
	comp := compiler.New()
	comp.BaseDir = dir
	if err := comp.Compile(program); err != nil {
		return nil, err, nil
	}
	machine := New(comp.Bytecode())
	return machine, nil, machine.Run()
}

// requirePrivateFailure accepts either a Go error from Run (OpImport pushes
// the Error and also returns it as a Go error, vm.go:436-441) or an *obj.Error
// value left on the stack (OpGetProp only pushes, vm.go:445-460), as long as
// the message mentions "private".
func requirePrivateFailure(t *testing.T, runErr error, machine *VM) {
	t.Helper()
	if runErr != nil {
		if !strings.Contains(runErr.Error(), "private") {
			t.Fatalf("expected error mentioning private, got: %s", runErr.Error())
		}
		return
	}
	if machine == nil {
		t.Fatalf("expected private failure, got no machine and no error")
	}
	errObj, ok := machine.LastPoppedStackElem().(*obj.Error)
	if !ok {
		t.Fatalf("expected Error object, got=%T (%+v)",
			machine.LastPoppedStackElem(), machine.LastPoppedStackElem())
	}
	if !strings.Contains(errObj.Message, "private") {
		t.Fatalf("expected error mentioning private, got: %q", errObj.Message)
	}
}

func TestImportStarPublicBinding(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"version", `from "` + mod + `" import *; version;`, 1},
		{"greetCall", `from "` + mod + `" import *; greet("world");`, "hello, world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, compileErr, runErr := runVMWithBaseDir(t, dir, tt.input)
			if compileErr != nil {
				t.Fatalf("compile error: %s", compileErr)
			}
			if runErr != nil {
				t.Fatalf("vm error: %s", runErr)
			}
			testExpectedObject(t, tt.expected, machine.LastPoppedStackElem())
		})
	}
}

func TestImportNamedPublicBinding(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"version", `from "` + mod + `" import version; version;`, 1},
		{"greetCall", `from "` + mod + `" import greet; greet("named");`, "hello, named"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, compileErr, runErr := runVMWithBaseDir(t, dir, tt.input)
			if compileErr != nil {
				t.Fatalf("compile error: %s", compileErr)
			}
			if runErr != nil {
				t.Fatalf("vm error: %s", runErr)
			}
			testExpectedObject(t, tt.expected, machine.LastPoppedStackElem())
		})
	}
}

func TestImportNamedPrivateError(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name  string
		input string
	}{
		{"secret", `from "` + mod + `" import secret;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, compileErr, runErr := runVMWithBaseDir(t, dir, tt.input)
			if compileErr != nil {
				t.Fatalf("expected runtime private error, got compile error: %s", compileErr)
			}
			requirePrivateFailure(t, runErr, machine)
		})
	}
}

func TestImportModuleAliasProperty(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"version", `import "` + mod + `" as m; m.version;`, 1},
		{"greetCall", `import "` + mod + `" as m; m.greet("alias");`, "hello, alias"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, compileErr, runErr := runVMWithBaseDir(t, dir, tt.input)
			if compileErr != nil {
				t.Fatalf("compile error: %s", compileErr)
			}
			if runErr != nil {
				t.Fatalf("vm error: %s", runErr)
			}
			testExpectedObject(t, tt.expected, machine.LastPoppedStackElem())
		})
	}
}

func TestImportModulePrivateProperty(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name  string
		input string
	}{
		{"secret", `import "` + mod + `" as m; m.secret;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, compileErr, runErr := runVMWithBaseDir(t, dir, tt.input)
			if compileErr != nil {
				t.Fatalf("expected runtime private error, got compile error: %s", compileErr)
			}
			requirePrivateFailure(t, runErr, machine)
		})
	}
}

// TestImportNestedTopLevelOnly documents the VM-only restriction from
// compiler.go:1171 ("import only supported at top-level in VM"): an import
// nested in a function body must fail at compile time.
func TestImportNestedTopLevelOnly(t *testing.T) {
	dir, mod := writeVMImportFixture(t)

	tests := []struct {
		name  string
		input string
	}{
		{"star", `fn() { from "` + mod + `" import *; };`},
		{"named", `fn() { from "` + mod + `" import version; };`},
		{"alias", `fn() { import "` + mod + `" as m; };`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, compileErr, _ := runVMWithBaseDir(t, dir, tt.input)
			if compileErr == nil {
				t.Fatalf("expected top-level compile error, got none")
			}
			if !strings.Contains(compileErr.Error(), "top-level") {
				t.Fatalf("expected error mentioning top-level, got: %s", compileErr.Error())
			}
		})
	}
}
