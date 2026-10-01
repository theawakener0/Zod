package compiler

import (
	zstdlib "github.com/theawakener0/Zod/stdlib"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveStdModuleEmbeddedFallback(t *testing.T) {
	t.Setenv("ZOD_STDLIB", "")
	t.Chdir(t.TempDir())

	p, err := ResolveModulePath("", "std/fmt")
	if err != nil {
		t.Fatalf("ResolveModulePath(%q) error: %v", "std/fmt", err)
	}
	if p != "zod-embedded/fmt.zd" {
		t.Fatalf("ResolveModulePath(%q) = %q, want %q", "std/fmt", p, "zod-embedded/fmt.zd")
	}

	src, err := zstdlib.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", p, err)
	}
	if !strings.Contains(string(src), "println") {
		t.Errorf("embedded fmt.zd missing %q, got:\n%s", "println", src)
	}
}

func TestResolveStdModuleDiskWins(t *testing.T) {
	root := t.TempDir()
	stdDir := filepath.Join(root, "stdlib")
	if err := os.MkdirAll(stdDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error: %v", stdDir, err)
	}
	distinct := "pub println := fn(x) { __print(\"disk\") }\n"
	if err := os.WriteFile(filepath.Join(stdDir, "fmt.zd"), []byte(distinct), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error: %v", work, err)
	}

	t.Setenv("ZOD_STDLIB", "")
	t.Chdir(work)

	p, err := ResolveModulePath("", "std/fmt")
	if err != nil {
		t.Fatalf("ResolveModulePath(%q) error: %v", "std/fmt", err)
	}
	if strings.HasPrefix(p, zstdlib.VirtualPrefix) {
		t.Fatalf("ResolveModulePath(%q) = %q, want a real on-disk path, not virtual", "std/fmt", p)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error: %v", p, err)
	}
	if string(got) != distinct {
		t.Errorf("read %q = %q, want %q", p, got, distinct)
	}
}

func TestResolveStdModuleMissingStillFails(t *testing.T) {
	t.Setenv("ZOD_STDLIB", "")
	t.Chdir(t.TempDir())

	if _, err := ResolveModulePath("", "std/no_such_module_xyz"); err == nil {
		t.Fatalf("ResolveModulePath(%q) err = nil, want error", "std/no_such_module_xyz")
	}
}
