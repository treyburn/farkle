package main

import "testing"

func TestBuildVersion(t *testing.T) {
	// The value depends on how the test binary was produced - under `go test`
	// the build info is always present, so only the absence of an empty string
	// is worth asserting.
	if got := buildVersion(); got == "" {
		t.Error("buildVersion() = \"\", want a non-empty version")
	}
}
