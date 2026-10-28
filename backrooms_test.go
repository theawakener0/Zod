package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theawakener0/Zod/repl"
)

// TestBackroomsZodSuite runs the demo's Zod test suite as part of
// `go test ./...`, so world, DDA, collision, input, palette, presenter,
// and mouse-parser regressions are caught by the normal test command.
func TestBackroomsZodSuite(t *testing.T) {
	path := filepath.Join("examples", "demo", "backrooms", "tests.zd")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := repl.Execute(string(source), os.Stdout, filepath.Dir(path)); err != nil {
		t.Fatalf("backrooms Zod suite failed: %v", err)
	}
}
