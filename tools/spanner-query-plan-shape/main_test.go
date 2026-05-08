package main

import (
	"testing"
)

func TestIsRawPlanOutput(t *testing.T) {
	for _, output := range []string{"json", "yaml"} {
		if !isRawPlanOutput(output) {
			t.Fatalf("isRawPlanOutput(%q) = false, want true", output)
		}
	}
	for _, output := range []string{"nodes", "reference", "textproto"} {
		if isRawPlanOutput(output) {
			t.Fatalf("isRawPlanOutput(%q) = true, want false", output)
		}
	}
}
