//go:build !integration

package main

import (
	"os"
	"runtime/debug"
	"testing"
)

func TestMain(m *testing.M) {
	// The query-gen CLI tests build many short-lived GoogleSQL frontend
	// catalogs. go-googlesql v0.2.0 still has WASM finalizer timing issues, so
	// keep GC disabled for this test process just like the root analyzer tests.
	oldGCPercent := debug.SetGCPercent(-1)
	code := m.Run()
	debug.SetGCPercent(oldGCPercent)
	os.Exit(code)
}
