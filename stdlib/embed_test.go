package stdlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveEmbedded(t *testing.T) {
	got, ok := Resolve("fmt")
	if !ok {
		t.Fatalf("Resolve(%q) failed, want ok=true", "fmt")
	}
	if got != "zod-embedded/fmt.zd" {
		t.Errorf("Resolve(%q) = %q, want %q", "fmt", got, "zod-embedded/fmt.zd")
	}

	got, ok = Resolve("fmt.zd")
	if !ok {
		t.Fatalf("Resolve(%q) failed, want ok=true", "fmt.zd")
	}
	if got != "zod-embedded/fmt.zd" {
		t.Errorf("Resolve(%q) = %q, want %q", "fmt.zd", got, "zod-embedded/fmt.zd")
	}

	if _, ok := Resolve("no-such-module"); ok {
		t.Errorf("Resolve(%q) ok=true, want ok=false", "no-such-module")
	}
	if _, ok := Resolve(""); ok {
		t.Errorf("Resolve(%q) ok=true, want ok=false", "")
	}
	if _, ok := Resolve("../escape"); ok {
		t.Errorf("Resolve(%q) ok=true, want ok=false", "../escape")
	}
	if _, ok := Resolve("/abs/path"); ok {
		t.Errorf("Resolve(%q) ok=true, want ok=false", "/abs/path")
	}
}

func TestResolveAllStdlibModules(t *testing.T) {
	modules := []string{
		"anim", "array", "bytes", "canvas", "color", "crypto", "fmt", "fs",
		"hash", "http", "io", "iter", "json", "log", "math", "matrix", "os",
		"path", "rand", "sort", "str", "term", "test", "time", "vec",
	}
	for _, m := range modules {
		got, ok := Resolve(m)
		if !ok {
			t.Errorf("Resolve(%q) failed, want ok=true", m)
			continue
		}
		if want := VirtualPrefix + m + ".zd"; got != want {
			t.Errorf("Resolve(%q) = %q, want %q", m, got, want)
		}
	}
}

func TestReadFileVirtualAndDisk(t *testing.T) {
	src, err := ReadFile("zod-embedded/fmt.zd")
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", "zod-embedded/fmt.zd", err)
	}
	if len(src) == 0 {
		t.Fatalf("ReadFile(%q) returned empty content", "zod-embedded/fmt.zd")
	}
	if !strings.Contains(string(src), "println") {
		t.Errorf("embedded fmt.zd missing %q, got:\n%s", "println", src)
	}

	if !IsVirtual("zod-embedded/fmt.zd") {
		t.Errorf("IsVirtual(%q) = false, want true", "zod-embedded/fmt.zd")
	}
	if IsVirtual("/tmp/x.zd") {
		t.Errorf("IsVirtual(%q) = true, want false", "/tmp/x.zd")
	}

	file := filepath.Join(t.TempDir(), "passthrough.zd")
	content := "pub hello := fn() { __print(\"disk passthrough\") }\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error: %v", file, err)
	}
	got, err := ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", file, err)
	}
	if string(got) != content {
		t.Errorf("ReadFile(%q) = %q, want %q", file, got, content)
	}
}

func TestReadFileVirtualNestedImport(t *testing.T) {
	src, err := ReadFile("zod-embedded/vec.zd")
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", "zod-embedded/vec.zd", err)
	}
	if !strings.Contains(string(src), `import "std/math"`) {
		t.Errorf("embedded vec.zd missing nested import %q, got:\n%s", `import "std/math"`, src)
	}
}
