package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultConfigFile  = "docs.json"
	defaultSidebarFile = "sidebar.json"
	defaultSrcDir      = "docs"
	defaultOutDir      = "dist"
	defaultPublicDir   = "public"
	defaultTitle       = "DOCS_GO"
	defaultDescription = "Documentation built with docs-go."
	defaultLang        = "en"
)

type Config struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Lang        string `json:"lang"`
	Base        string `json:"base"`

	SrcDir    string `json:"src_dir"`
	OutDir    string `json:"out_dir"`
	PublicDir string `json:"public_dir"`

	CleanURLs   bool   `json:"clean_urls"`
	SiteURL     string `json:"site_url"`
	SidebarFile string `json:"sidebar_file"`
	Sitemap     bool   `json:"sitemap"`

	Logo       string `json:"logo"`
	Favicon    string `json:"favicon"`
	Appearance string `json:"appearance"`

	Nav         []NavItem    `json:"nav"`
	SocialLinks []SocialLink `json:"social_links"`
	UI          UIConfig     `json:"ui"`

	Outline     OutlineConfig     `json:"outline"`
	Containers  map[string]string `json:"containers"`
	EditLink    string            `json:"edit_link"`
	LastUpdated bool              `json:"last_updated"`
	PrevNext    bool              `json:"prev_next"`
	PrevText    string            `json:"prev_text"`
	NextText    string            `json:"next_text"`
}

type NavItem struct {
	Text        string `json:"text"`
	Link        string `json:"link"`
	ActiveMatch string `json:"active_match"`
}

type SocialLink struct {
	Icon  string `json:"icon"`
	Link  string `json:"link"`
	Label string `json:"label"`
}

type UIConfig struct {
	SkipToContent string `json:"skip_to_content"`
	Menu          string `json:"menu"`
	Theme         string `json:"theme"`
	EditPage      string `json:"edit_page"`
	LastUpdated   string `json:"last_updated"`
	NotFoundTitle string `json:"not_found_title"`
	NotFoundText  string `json:"not_found_text"`
	BackHome      string `json:"back_home"`
	Copy          string `json:"copy"`
	Copied        string `json:"copied"`
}

type OutlineConfig struct {
	Label string `json:"label"`
	Level []int  `json:"level"`
}

func defaults() Config {
	return Config{
		Title:       defaultTitle,
		Description: defaultDescription,
		Lang:        defaultLang,
		Base:        "/",
		SrcDir:      defaultSrcDir,
		OutDir:      defaultOutDir,
		PublicDir:   defaultPublicDir,
		CleanURLs:   true,
		SidebarFile: defaultSidebarFile,
		Appearance:  "auto",
		Outline:     OutlineConfig{Label: "On this page", Level: []int{2, 3}},
		Containers:  defaultContainers(),
		LastUpdated: true,
		PrevNext:    true,
		PrevText:    "Previous page",
		NextText:    "Next page",
		UI: UIConfig{
			SkipToContent: "Skip to content",
			Menu:          "Menu",
			Theme:         "Toggle theme",
			EditPage:      "Edit this page",
			LastUpdated:   "Last updated",
			NotFoundTitle: "Page not found",
			NotFoundText:  "The page you are looking for does not exist or has moved.",
			BackHome:      "Back to home",
			Copy:          "Copy",
			Copied:        "Copied",
		},
	}
}

func loadConfig(projectRoot string, configFile string) (Config, string, error) {
	if configFile == "" {
		configFile = defaultConfigFile
	}

	path := configFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return Config{}, "", fmt.Errorf("config path: %w", err)
	}

	data, err := os.ReadFile(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, absolute, fmt.Errorf("no %s: run `docs-go init` first", filepath.Base(absolute))
		}
		return Config{}, absolute, fmt.Errorf("read config: %w", err)
	}

	config := defaults()
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, absolute, fmt.Errorf("parse %s: %w", absolute, err)
	}

	config.normalize()
	if err := config.validate(); err != nil {
		return Config{}, absolute, fmt.Errorf("%s: %w", absolute, err)
	}

	return config, absolute, nil
}

func (config *Config) normalize() {
	config.Title = firstNonEmpty(config.Title, defaultTitle)
	config.Description = firstNonEmpty(config.Description, defaultDescription)
	config.Lang = firstNonEmpty(config.Lang, defaultLang)
	config.SrcDir = firstNonEmpty(config.SrcDir, defaultSrcDir)
	config.OutDir = firstNonEmpty(config.OutDir, defaultOutDir)
	config.PublicDir = firstNonEmpty(config.PublicDir, defaultPublicDir)
	config.SidebarFile = firstNonEmpty(config.SidebarFile, defaultSidebarFile)

	config.Base = normalizeBase(config.Base)
	config.SiteURL = strings.TrimRight(strings.TrimSpace(config.SiteURL), "/")
	config.Logo = strings.TrimSpace(config.Logo)
	config.Favicon = strings.TrimSpace(config.Favicon)
	config.EditLink = strings.TrimSpace(config.EditLink)

	if config.Appearance != "light" && config.Appearance != "dark" {
		config.Appearance = "auto"
	}

	config.Outline.Label = firstNonEmpty(config.Outline.Label, "On this page")
	config.Outline.Level = normalizeLevels(config.Outline.Level)

	if config.Containers == nil {
		config.Containers = make(map[string]string)
	}
	for name, label := range defaultContainers() {
		if strings.TrimSpace(config.Containers[name]) == "" {
			config.Containers[name] = label
		}
	}

	config.PrevText = firstNonEmpty(config.PrevText, "Previous page")
	config.NextText = firstNonEmpty(config.NextText, "Next page")

	defaultsForUI(&config.UI)
}

func (config Config) validate() error {
	if strings.Contains(config.Base, "..") || strings.ContainsAny(config.Base, "?#") {
		return fmt.Errorf("base must be a plain path such as / or /docs/")
	}

	if config.SiteURL != "" {
		parsed, err := url.Parse(config.SiteURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("site_url must be an absolute URL")
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("site_url must start with http:// or https://")
		}
	}

	if config.Sitemap && config.SiteURL == "" {
		return fmt.Errorf("sitemap requires site_url")
	}

	return nil
}

func defaultsForUI(ui *UIConfig) {
	ui.SkipToContent = firstNonEmpty(ui.SkipToContent, "Skip to content")
	ui.Menu = firstNonEmpty(ui.Menu, "Menu")
	ui.Theme = firstNonEmpty(ui.Theme, "Toggle theme")
	ui.EditPage = firstNonEmpty(ui.EditPage, "Edit this page")
	ui.LastUpdated = firstNonEmpty(ui.LastUpdated, "Last updated")
	ui.NotFoundTitle = firstNonEmpty(ui.NotFoundTitle, "Page not found")
	ui.NotFoundText = firstNonEmpty(ui.NotFoundText, "The page you are looking for does not exist or has moved.")
	ui.BackHome = firstNonEmpty(ui.BackHome, "Back to home")
	ui.Copy = firstNonEmpty(ui.Copy, "Copy")
	ui.Copied = firstNonEmpty(ui.Copied, "Copied")
}

func defaultContainers() map[string]string {
	return map[string]string{
		"tip": "TIP", "info": "INFO", "note": "NOTE",
		"warning": "WARNING", "danger": "DANGER", "details": "Details",
	}
}

func normalizeLevels(levels []int) []int {
	filtered := make([]int, 0, len(levels))
	for _, level := range levels {
		if level >= 1 && level <= 6 {
			filtered = append(filtered, level)
		}
	}
	if len(filtered) == 0 {
		return []int{2, 3}
	}
	return filtered
}

func normalizeBase(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	value = path.Clean(value)
	if value == "." {
		return "/"
	}
	if !strings.HasSuffix(value, "/") {
		value += "/"
	}
	return value
}

func firstNonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func resolveConfigPath(root string, value string) string {
	if value == "" {
		return root
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Join(root, value)
}

func applyConfigOverrides(config *Config, src string, out string) {
	if src != "" {
		config.SrcDir = src
	}
	if out != "" {
		config.OutDir = out
	}
}
