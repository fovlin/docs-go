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

func loadSidebar(file string) (*SidebarConfig, error) {
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

	raw := make(map[string][]SidebarItem)
	switch trimmed[0] {
	case '[':
		var items []SidebarItem
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, fmt.Errorf("parse sidebar %s: %w", file, err)
		}
		raw["/"] = items

	case '{':
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			return nil, fmt.Errorf("parse sidebar %s: %w", file, err)
		}
	default:
		return nil, fmt.Errorf("parse sidebar %s: expected an array or an object", file)
	}

	for key, items := range raw {
		prefix := normalizeSidebarPrefix(key)
		if _, exists := config.Sections[prefix]; exists {
			return nil, fmt.Errorf("sidebar prefix %s is defined more than once", prefix)
		}
		config.Sections[prefix] = items
		config.keys = append(config.keys, prefix)
	}
	sort.Slice(config.keys, func(i, j int) bool {
		if len(config.keys[i]) != len(config.keys[j]) {
			return len(config.keys[i]) > len(config.keys[j])
		}
		return config.keys[i] < config.keys[j]
	})

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

func normalizeSidebarPrefix(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "/"
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
