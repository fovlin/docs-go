package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

// SidebarItem is one entry in sidebar.json.
type SidebarItem struct {
	Text      string        `json:"text"`
	Link      string        `json:"link"`
	Collapsed bool          `json:"collapsed"`
	Items     []SidebarItem `json:"items"`
}

// SidebarConfig accepts either one array for every page or an object keyed by
// URL prefix.
type SidebarConfig struct {
	Global   []SidebarItem
	Sections map[string][]SidebarItem
	keys     []string
	file     string
	found    bool
}

// SidebarNode is the resolved tree consumed by the layout.
type SidebarNode struct {
	Text      string
	Href      string
	Path      string
	External  bool
	Active    bool
	HasActive bool
	Collapsed bool
	Open      bool
	Items     []*SidebarNode
}

func loadSidebar(file string, base string) (*SidebarConfig, error) {
	config := &SidebarConfig{
		Sections: make(map[string][]SidebarItem),
		file:     file,
	}

	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("read sidebar %s: %w", file, err)
	}

	config.found = true
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return config, nil
	}

	switch trimmed[0] {
	case '[':
		if err := json.Unmarshal(trimmed, &config.Global); err != nil {
			return nil, fmt.Errorf("parse sidebar %s: %w", file, err)
		}
	case '{':
		raw := make(map[string][]SidebarItem)
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			return nil, fmt.Errorf("parse sidebar %s: %w", file, err)
		}
		for key, items := range raw {
			prefix := normalizeSidebarPrefix(key, base)
			config.Sections[prefix] = items
			config.keys = append(config.keys, prefix)
		}
		sort.Slice(config.keys, func(i, j int) bool {
			if len(config.keys[i]) != len(config.keys[j]) {
				return len(config.keys[i]) > len(config.keys[j])
			}
			return config.keys[i] < config.keys[j]
		})
	default:
		return nil, fmt.Errorf("parse sidebar %s: expected an array or an object", file)
	}

	return config, nil
}

func sidebarPrefixMatches(prefix string, pageURL string) bool {
	if prefix == "/" {
		return true
	}
	if pageURL == prefix {
		return true
	}
	return strings.HasPrefix(pageURL, strings.TrimSuffix(prefix, "/")+"/")
}

func normalizeSidebarPrefix(key string, base string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "/"
	}

	if base != "/" && strings.HasPrefix(key, base) {
		key = strings.TrimPrefix(key, base)
		key = "/" + strings.TrimPrefix(key, "/")
	}
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}

	key = path.Clean(key)
	if key == "." {
		return "/"
	}
	return key
}

func (app *App) matchedSidebar(pageURL string) []*SidebarNode {
	for _, key := range app.Sidebar.keys {
		if sidebarPrefixMatches(key, pageURL) {
			return app.SidebarTrees[key]
		}
	}

	return nil
}
