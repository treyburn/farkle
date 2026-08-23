package main

import "testing"

func TestBuildVersion(t *testing.T) {
	t.Run("uses the stamped version when present", func(t *testing.T) {
		orig := version
		t.Cleanup(func() { version = orig })

		version = "v1.2.3"
		if got := buildVersion(); got != "v1.2.3" {
			t.Errorf("buildVersion() = %q, want %q", got, "v1.2.3")
		}
	})

	t.Run("falls back when the version was not stamped", func(t *testing.T) {
		orig := version
		t.Cleanup(func() { version = orig })

		// Without ldflags the value comes from the build info, which varies by
		// how the binary was produced - only the absence of an empty string is
		// worth asserting.
		version = ""
		if got := buildVersion(); got == "" {
			t.Error("buildVersion() = \"\", want a non-empty fallback")
		}
	})
}
