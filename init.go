package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func runInit(root string, force bool) error {
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		return fmt.Errorf("not a directory: %s", root)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	configFile := filepath.Join(root, defaultConfigFile)
	if _, err := os.Stat(configFile); err == nil && !force {
		return fmt.Errorf("%s already exists; use --force", configFile)
	}

	var created, kept []string

	err := fs.WalkDir(themeFS, "scaffold", func(current string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		relative := strings.TrimPrefix(current, "scaffold/")
		target := filepath.Join(root, filepath.FromSlash(relative))
		if _, err := os.Stat(target); err == nil && !force {
			kept = append(kept, relative)
			return nil
		}

		data, err := themeFS.ReadFile(current)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", relative, err)
		}

		created = append(created, relative)
		return nil
	})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(root, defaultPublicDir), 0o755); err != nil {
		return fmt.Errorf("create public: %w", err)
	}

	display := displayPath(root)
	logInfo("initialized %s", display)
	for _, name := range created {
		logPlain("  create  %s\n", name)
	}
	for _, name := range kept {
		logPlain("  keep    %s\n", name)
	}
	logPlain("\nnext:\n  cd %s\n  docs-go dev\n", display)

	return nil
}

func displayPath(root string) string {
	cwd, _ := os.Getwd()
	relative, err := filepath.Rel(cwd, root)
	if err == nil && !strings.HasPrefix(relative, "..") {
		return relative
	}
	return root
}
