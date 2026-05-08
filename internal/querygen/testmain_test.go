package querygen

import (
	"os"
	"runtime/debug"
	"testing"
)

func TestMain(m *testing.M) {
	// Querygen tests construct many short-lived GoogleSQL frontend objects; keep
	// GC disabled for this test process to avoid WASM finalizer timing issues.
	oldGCPercent := debug.SetGCPercent(-1)
	code := m.Run()
	debug.SetGCPercent(oldGCPercent)
	os.Exit(code)
}
