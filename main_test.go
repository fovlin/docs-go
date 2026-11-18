package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCommand(t *testing.T) {
	if cmd, _ := resolveCommand(nil); cmd != "build" {
		t.Fatalf("default command = %q", cmd)
	}
	if cmd, _ := resolveCommand([]string{"dev"}); cmd != "dev" {
		t.Fatalf("dev command = %q", cmd)
	}
	if cmd, _ := resolveCommand([]string{"."}); cmd != "build" {
		t.Fatalf("project path command = %q", cmd)
	}
	if cmd, _ := resolveCommand([]string{"no-such-command"}); cmd != "unknown" {
		t.Fatalf("unknown command = %q", cmd)
	}
}

func TestValidateFlagValues(t *testing.T) {
	if err := validateFlagValues([]string{"build", "-c", "config/docs.json"}); err != nil {
		t.Fatal(err)
	}
	if err := validateFlagValues([]string{"build", "-c"}); err == nil {
		t.Fatal("expected missing value error")
	}
}

func TestInitCustomConfig(t *testing.T) {
	root := t.TempDir()
	if err := runInit(root, false, "site.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "site.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "config/docs.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected config/docs.json: %v", err)
	}
	app, err := LoadApp(root, "site.json", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if app.Config.Title == "" {
		t.Fatal("empty title")
	}
}
