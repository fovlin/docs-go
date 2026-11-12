package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// App is one immutable snapshot of a docs project. Everything a render needs
// is resolved before the snapshot is published; requests never mutate it.
type App struct {
	Config

	Root        string
	SrcDir      string
	OutDir      string
	PublicDir   string
	ConfigPath  string
	SidebarPath string

	Sidebar *SidebarConfig

	URLTab       map[string]NavInfo
	Nav          []*NavLink
	SidebarTrees map[string][]*SidebarNode

	Layout       []layoutPart
	Favicon      []byte
	AssetVersion string
	Bootstrap    string

	Titles *TitleCache
}

// LoadApp reads one coherent project snapshot. The returned App is fully
// resolved and must not be modified by callers.
func LoadApp(projectRoot, configFile, srcFlag, outFlag string) (*App, error) {
	absoluteRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
	}

	loaded, path, err := loadConfig(absoluteRoot, configFile)
	if err != nil {
		return nil, err
	}
	applyConfigOverrides(&loaded, srcFlag, outFlag)

	app := &App{
		Config:      loaded,
		Root:        absoluteRoot,
		ConfigPath:  path,
		SrcDir:      resolveConfigPath(absoluteRoot, loaded.SrcDir),
		OutDir:      resolveConfigPath(absoluteRoot, loaded.OutDir),
		PublicDir:   resolveConfigPath(absoluteRoot, loaded.PublicDir),
		SidebarPath: resolveConfigPath(absoluteRoot, loaded.SidebarFile),
		Titles:      newTitleCache(),
	}

	info, err := os.Stat(app.SrcDir)
	if err != nil {
		return nil, fmt.Errorf("source directory %s: %w", app.SrcDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source directory is not a directory: %s", app.SrcDir)
	}

	app.Sidebar, err = loadSidebar(app.SidebarPath, app.Base)
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

	payload, err := json.Marshal(struct {
		Base string   `json:"base"`
		UI   UIConfig `json:"ui"`
	}{
		Base: app.Base,
		UI:   app.UI,
	})
	if err != nil {
		return nil, fmt.Errorf("encode bootstrap: %w", err)
	}
	app.Bootstrap = string(payload)

	app.Nav = app.resolveNav()
	app.SidebarTrees = app.resolveSidebarTrees()

	if err := app.buildURLTab(); err != nil {
		return nil, err
	}

	return app, nil
}
