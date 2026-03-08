package utils

import (
	"os"
	"testing"
)

func TestGetEnvVar(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "hello")
		defer os.Unsetenv("TEST_ENV_VAR")

		got := GetEnvVar("TEST_ENV_VAR", false)
		if got != "hello" {
			t.Errorf("GetEnvVar() = %q, want %q", got, "hello")
		}
	})

	t.Run("returns empty string when unset and panic is false", func(t *testing.T) {
		os.Unsetenv("TEST_ENV_MISSING")

		got := GetEnvVar("TEST_ENV_MISSING", false)
		if got != "" {
			t.Errorf("GetEnvVar() = %q, want empty", got)
		}
	})

	t.Run("panics when unset and panic is true", func(t *testing.T) {
		os.Unsetenv("TEST_ENV_MISSING")

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic, got none")
			}
		}()

		GetEnvVar("TEST_ENV_MISSING", true)
	})

	t.Run("returns value when set and panic is true", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR2", "world")
		defer os.Unsetenv("TEST_ENV_VAR2")

		got := GetEnvVar("TEST_ENV_VAR2", true)
		if got != "world" {
			t.Errorf("GetEnvVar() = %q, want %q", got, "world")
		}
	})
}
