package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

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
