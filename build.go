package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runBuild() error {
	if err := validateOutput(outDir); err != nil {
		return err
	}

	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear output directory: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := copyTree(srcDir, outDir, true); err != nil {
		return fmt.Errorf("copy source assets: %w", err)
	}

	if err := writeThemeAssets(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(outDir, "favicon.ico"), favicon, 0o644); err != nil {
		return fmt.Errorf("write favicon: %w", err)
	}

	count, err := writePages()
	if err != nil {
		return err
	}

	notFound := notFoundPage()
	document := renderLayout(layout, notFound)
	if err := writeOutputFile(outDir, "404.html", document); err != nil {
		return fmt.Errorf("write 404 page: %w", err)
	}

	if config.Sitemap {
		if err := writeSitemap(); err != nil {
			return err
		}
	}

	if _, err := os.Stat(publicDir); err == nil {
		if err := copyTree(publicDir, outDir, false); err != nil {
			return fmt.Errorf("copy public directory: %w", err)
		}
	}

	logInfo("built %d pages in %s", count, time.Since(buildStart).Round(time.Millisecond))
	logInfo("output: %s", outDir)

	return nil
}

func writePages() (int, error) {
	count := 0

	err := walkPages(func(page *Page) error {
		if err := renderPage(page); err != nil {
			return fmt.Errorf("render %s: %w", page.SourcePath, err)
		}

		count++
		return nil
	})

	return count, err
}

func walkPages(visit func(*Page) error) error {
	err := filepath.WalkDir(srcDir, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if current == srcDir {
				return nil
			}

			name := entry.Name()
			if name == ".git" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}

			return nil
		}

		if !isMarkdownFile(current) {
			return nil
		}

		relative, err := filepath.Rel(srcDir, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		page, err := loadPage(relative)
		if err != nil {
			return fmt.Errorf("read %s: %w", relative, err)
		}

		return visit(page)
	})

	return err
}

func validateOutput(output string) error {
	output = filepath.Clean(output)
	root = filepath.Clean(root)
	srcDir = filepath.Clean(srcDir)
	publicDir = filepath.Clean(publicDir)

	if output == root || isInside(output, root) {
		return fmt.Errorf("out_dir must not be the project root or contain it")
	}

	for _, directory := range []string{srcDir, publicDir} {
		if output == directory || isInside(output, directory) || isInside(directory, output) {
			return fmt.Errorf("out_dir must not overlap %s", directory)
		}
	}

	if info, err := os.Lstat(output); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("out_dir must not be a symlink: %s", output)
		}
		if !info.IsDir() {
			return fmt.Errorf("out_dir exists and is not a directory: %s", output)
		}
	}

	return nil
}

func isInside(parent string, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)
	if parent == child {
		return false
	}

	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func writeThemeAssets() error {
	assets := filepath.Join(outDir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}

	for _, name := range []string{"style.css", "app.js", "init.js"} {
		data, err := themeFS.ReadFile("theme/" + name)
		if err != nil {
			return fmt.Errorf("read theme asset %s: %w", name, err)
		}

		if err := os.WriteFile(filepath.Join(assets, name), data, 0o644); err != nil {
			return fmt.Errorf("write theme asset %s: %w", name, err)
		}
	}

	return nil
}

func themeVersion() string {
	hash := sha256.New()
	for _, name := range []string{"style.css", "app.js", "init.js"} {
		data, _ := themeFS.ReadFile("theme/" + name)
		hash.Write(data)
	}
	hash.Write(favicon)

	return hex.EncodeToString(hash.Sum(nil))[:8]
}

func sitemapBytes() ([]byte, error) {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	builder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	if err := walkPages(func(page *Page) error {
		builder.WriteString("  <url>\n    <loc>")
		_ = xml.EscapeText(&builder, []byte(config.SiteURL+page.Href))
		builder.WriteString("</loc>\n")
		if !page.ModTime.IsZero() {
			builder.WriteString("    <lastmod>")
			builder.WriteString(page.ModTime.Format("2006-01-02"))
			builder.WriteString("</lastmod>\n")
		}
		builder.WriteString("  </url>\n")
		return nil
	}); err != nil {
		return nil, err
	}

	builder.WriteString("</urlset>\n")

	return []byte(builder.String()), nil
}

func writeSitemap() error {
	document, err := sitemapBytes()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(outDir, "sitemap.xml"), document, 0o644); err != nil {
		return fmt.Errorf("write sitemap: %w", err)
	}

	return nil
}

func writeOutputFile(directory string, relative string, data []byte) error {
	target := filepath.Join(directory, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, data, 0o644)
}

func copyTree(source string, destination string, skipMarkdown bool) error {
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, current)
		if err != nil || relative == "." {
			return err
		}

		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}

			return os.MkdirAll(filepath.Join(destination, relative), 0o755)
		}

		if skipMarkdown && isMarkdownFile(current) {
			return nil
		}

		return copyFile(current, filepath.Join(destination, relative))
	})
}

func copyFile(source string, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(destination)
	if err != nil {
		return err
	}

	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}

	return output.Close()
}

func isMarkdownFile(name string) bool {
	extension := strings.ToLower(filepath.Ext(name))
	return extension == ".md" || extension == ".markdown"
}
