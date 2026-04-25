package connector_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMain roots the process at the services/refinery directory before any test runs.
// This ensures config.InitDefaults() can resolve "configs/protected_entities.json"
// via its relative path, regardless of which directory `go test` is invoked from.
func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile is .../apps/sombra/tests/testmain_test.go
	// Walk up 3 dirs (tests/ → sombra/ → apps/ → dev/) then down to services/refinery
	refineryRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "services", "refinery")
	if err := os.Chdir(refineryRoot); err != nil {
		panic("TestMain: could not chdir to refinery root: " + err.Error())
	}
	os.Exit(m.Run())
}
