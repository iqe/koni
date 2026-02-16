package main

import (
	"testing"

	toml "github.com/pelletier/go-toml"
)

func TestGetConfigValueDefault(t *testing.T) {
	tree, err := toml.Load(`
key1 = "value1"
number = 42
`)
	if err != nil {
		t.Fatalf("failed to parse TOML: %v", err)
	}

	t.Run("existing key", func(t *testing.T) {
		got := getConfigValueDefault(tree, "key1", "default")
		if got != "value1" {
			t.Errorf("got %q, want %q", got, "value1")
		}
	})

	t.Run("missing key returns default", func(t *testing.T) {
		got := getConfigValueDefault(tree, "missing", "fallback")
		if got != "fallback" {
			t.Errorf("got %q, want %q", got, "fallback")
		}
	})
}

func TestGetBoolConfigValueDefault(t *testing.T) {
	tree, err := toml.Load(`
enabled = "yes"
disabled = "no"
`)
	if err != nil {
		t.Fatalf("failed to parse TOML: %v", err)
	}

	t.Run("yes value", func(t *testing.T) {
		got := getBoolConfigValueDefault(tree, "enabled", false)
		if !got {
			t.Error("got false, want true")
		}
	})

	t.Run("no value", func(t *testing.T) {
		got := getBoolConfigValueDefault(tree, "disabled", true)
		if got {
			t.Error("got true, want false")
		}
	})

	t.Run("missing key uses default true", func(t *testing.T) {
		got := getBoolConfigValueDefault(tree, "missing", true)
		if !got {
			t.Error("got false, want true")
		}
	})

	t.Run("missing key uses default false", func(t *testing.T) {
		got := getBoolConfigValueDefault(tree, "missing", false)
		if got {
			t.Error("got true, want false")
		}
	})
}
