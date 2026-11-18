package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"acovia.net/docs-go/internal/record"
)

//go:embed theme scaffold
var themeFS embed.FS

const version = "1.0.2"


func runInit(root string, force bool, configName string) error {
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		return fmt.Errorf("not a directory: %s", root)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	if configName == "" {
		configName = defaultConfigFile
	}
	configTarget := configName
	if !filepath.IsAbs(configTarget) {
		configTarget = filepath.Join(root, configTarget)
	}
	if _, err := os.Stat(configTarget); err == nil && !force {
		return fmt.Errorf("%s already exists; use --force", configTarget)
	}

	var created, kept []string

	err := fs.WalkDir(themeFS, "scaffold", func(current string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		relative := strings.TrimPrefix(current, "scaffold/")
		target := filepath.Join(root, filepath.FromSlash(relative))
		if relative == defaultConfigFile {
			target = configTarget
		}
		if info, err := os.Stat(target); err == nil && !force {
			if info.IsDir() {
				return fmt.Errorf("scaffold target %s is a directory", target)
			}
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
	relative, err := filepath.Rel("./", root)
	record.Info("initialized %s", relative)
		for _, name := range created {
			record.Info("  create  %s\n", name)
		}
		for _, name := range kept {
			record.Info("  keep    %s\n", name)
		}
		record.Info("\nnext:\n  cd %s\n  docs-go dev\n", relative)

	return nil
}