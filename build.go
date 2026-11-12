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

func runBuild(app *App) error {
	start := time.Now()

	if err := app.validateOutput(); err != nil {
		return err
	}

	if err := os.RemoveAll(app.OutDir); err != nil {
		return fmt.Errorf("clear output directory: %w", err)
	}

	if err := os.MkdirAll(app.OutDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := app.copyTree(app.SrcDir, app.OutDir, true); err != nil {
		return fmt.Errorf("copy source assets: %w", err)
	}

	if err := app.writeThemeAssets(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(app.OutDir, "favicon.ico"), app.Favicon, 0o644); err != nil {
		return fmt.Errorf("write favicon: %w", err)
	}

	count, err := app.writePages()
	if err != nil {
		return err
	}

	notFound := app.notFoundPage()
	document := app.renderLayout(notFound)
	if err := app.writeOutputFile("404.html", document); err != nil {
		return fmt.Errorf("write 404 page: %w", err)
	}

	if app.Sitemap {
		if err := app.writeSitemap(); err != nil {
			return err
		}
	}

	if _, err := os.Stat(app.PublicDir); err == nil {
		if err := app.copyTree(app.PublicDir, app.OutDir, false); err != nil {
			return fmt.Errorf("copy public directory: %w", err)
		}
	}

	logInfo("built %d pages in %s", count, time.Since(start).Round(time.Millisecond))
	logInfo("output: %s", app.OutDir)

	return nil
}

func (app *App) writePages() (int, error) {
	count := 0

	err := app.walkPages(func(page *Page) error {
		if err := app.renderPage(page); err != nil {
			return fmt.Errorf("render %s: %w", page.SourcePath, err)
		}

		count++
		return nil
	})

	return count, err
}

func (app *App) walkPages(visit func(*Page) error) error {
	err := filepath.WalkDir(app.SrcDir, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if current == app.SrcDir {
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

		relative, err := filepath.Rel(app.SrcDir, current)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)

		page, err := app.loadPage(relative)
		if err != nil {
			return fmt.Errorf("read %s: %w", relative, err)
		}

		return visit(page)
	})

	return err
}

func (app *App) validateOutput() error {
	output := filepath.Clean(app.OutDir)
	root := filepath.Clean(app.Root)
	srcDir := filepath.Clean(app.SrcDir)
	publicDir := filepath.Clean(app.PublicDir)

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

func (app *App) writeThemeAssets() error {
	assets := filepath.Join(app.OutDir, "assets")
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

func (app *App) themeVersion() string {
	hash := sha256.New()
	for _, name := range []string{"style.css", "app.js", "init.js"} {
		data, _ := themeFS.ReadFile("theme/" + name)
		hash.Write(data)
	}
	hash.Write(app.Favicon)

	return hex.EncodeToString(hash.Sum(nil))[:8]
}

func (app *App) sitemapBytes() ([]byte, error) {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	builder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	if err := app.walkPages(func(page *Page) error {
		builder.WriteString("  <url>\n    <loc>")
		_ = xml.EscapeText(&builder, []byte(app.SiteURL+page.Href))
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

func (app *App) writeSitemap() error {
	document, err := app.sitemapBytes()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(app.OutDir, "sitemap.xml"), document, 0o644); err != nil {
		return fmt.Errorf("write sitemap: %w", err)
	}

	return nil
}

func (app *App) writeOutputFile(relative string, data []byte) error {
	target := filepath.Join(app.OutDir, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, data, 0o644)
}

func (app *App) copyTree(source string, destination string, skipMarkdown bool) error {
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
