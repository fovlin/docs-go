package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type App struct {
	Config

	Root        string
	SrcDir      string
	OutDir      string
	PublicDir   string
	ConfigPath  string
	SidebarPath string

	Sidebar *SidebarConfig

	Nav          []*NavLink
	SidebarTrees map[string][]*SidebarNode

	Layout       []layoutPart
	Favicon      []byte
	AssetVersion string
	Bootstrap    string
}

func LoadApp(projectRoot, configFile, srcFlag, outFlag string) (*App, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, err
	}

	loaded, path, err := loadConfig(root, configFile)
	if err != nil {
		return nil, err
	}
	applyConfigOverrides(&loaded, srcFlag, outFlag)

	app := &App{
		Config:      loaded,
		Root:        root,
		ConfigPath:  path,
		SrcDir:      fmtConfigPath(loaded.SrcDir),
		OutDir:      fmtConfigPath(loaded.OutDir),
		PublicDir:   fmtConfigPath(loaded.PublicDir),
		SidebarPath: fmtConfigPath(loaded.SidebarFile),
	}

	for _, dir := range []*string{&app.SrcDir, &app.OutDir, &app.PublicDir, &app.SidebarPath} {
		if !filepath.IsAbs(*dir) {
			*dir = filepath.Join(root, *dir)
		}
	}

	info, err := os.Stat(app.SrcDir)
	if err != nil {
		return nil, fmt.Errorf("source directory %s: %w", app.SrcDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source directory is not a directory: %s", app.SrcDir)
	}

	app.Sidebar, err = loadSidebar(app.SidebarPath)
	if err != nil {
		return nil, err
	}

	layoutData, err := themeFS.ReadFile("theme/layout.html")
	if err != nil {
		return nil, fmt.Errorf("read theme layout: %w", err)
	}
	app.Layout = compileLayout(string(layoutData))

	app.Favicon, err = themeFS.ReadFile("theme/favicon.ico")
	if err != nil {
		return nil, fmt.Errorf("read theme favicon: %w", err)
	}
	app.AssetVersion = app.themeVersion()

	payload, err := json.Marshal(app.UI)
	if err != nil {
		return nil, fmt.Errorf("encode bootstrap: %w", err)
	}
	app.Bootstrap = string(payload)

	app.Nav = app.resolveNav()
	app.SidebarTrees = app.resolveSidebarTrees()

	return app, nil
}
