package engine_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMain roots the process at the repository root before any test runs.
// This ensures that config.InitDefaults() can resolve "configs/protected_entities.json"
// via its relative path, regardless of which directory `go test` is invoked from.
func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile is .../pkg/engine/testmain_test.go — walk up two directories to repo root.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	if err := os.Chdir(repoRoot); err != nil {
		panic("TestMain: could not chdir to repo root: " + err.Error())
	}
	os.Exit(m.Run())
}
