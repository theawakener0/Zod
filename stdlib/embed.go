// Package stdlib exposes the Zod standard library as an embedded filesystem
// so the interpreter can resolve "std/..." imports even when no stdlib
// directory exists next to the executable (e.g. after `go install`).
//
// Lookup priority is decided by the callers: they search on-disk stdlib
// locations first and only fall back to this embedded copy. The embedded
// modules are addressed by virtual paths prefixed with VirtualPrefix; those
// paths must never be passed to filepath.Abs or os.Stat.
package stdlib

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"strings"
)

// files holds every stdlib module (*.zd, flat directory).
//
//go:embed *.zd
var files embed.FS

// VirtualPrefix marks module paths that live in the embedded stdlib.
// It uses forward slashes only and no colon, so filepath.Base and
// filepath.Dir behave correctly on every platform:
//
//	filepath.Base("zod-embedded/fmt.zd") -> "fmt.zd"
//	filepath.Dir("zod-embedded/fmt.zd")  -> "zod-embedded"
const VirtualPrefix = "zod-embedded/"

// Resolve looks up a stdlib module by its path relative to the stdlib root
// ("fmt", "fmt.zd", or a forward-slash subpath). On success it returns a
// virtual path suitable for ReadFile and for use as an import target.
func Resolve(rel string) (string, bool) {
	if rel == "" {
		return "", false
	}
	// Import paths are slash-separated; tolerate backslashes on Windows.
	rel = strings.ReplaceAll(rel, `\`, "/")
	names := []string{rel}
	if !strings.HasSuffix(rel, ".zd") {
		names = append(names, rel+".zd")
	}
	for _, name := range names {
		name = path.Clean(name)
		if !fs.ValidPath(name) {
			continue
		}
		if _, err := fs.Stat(files, name); err == nil {
			return VirtualPrefix + name, true
		}
	}
	return "", false
}

// IsVirtual reports whether p is an embedded-stdlib virtual path.
func IsVirtual(p string) bool {
	return strings.HasPrefix(p, VirtualPrefix)
}

// ReadFile reads a module's source. Virtual paths are served from the
// embedded filesystem; every other path behaves exactly like os.ReadFile.
func ReadFile(p string) ([]byte, error) {
	if IsVirtual(p) {
		return files.ReadFile(strings.TrimPrefix(p, VirtualPrefix))
	}
	return os.ReadFile(p)
}
