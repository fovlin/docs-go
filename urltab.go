package main

import (
	"fmt"
	"strings"
)

// NavInfo is the previous/next pair for one page.
type NavInfo struct {
	PrevHref string
	NextHref string
}

func (app *App) buildURLTab() error {
	table := make(map[string]NavInfo)

	if app.Sidebar != nil {
		if err := app.addURLInfos(table, "/", app.Sidebar.Global); err != nil {
			return err
		}

		for route, items := range app.Sidebar.Sections {
			if err := app.addURLInfos(table, route, items); err != nil {
				return err
			}
		}
	}

	app.URLTab = table

	return nil
}

func (app *App) addURLInfos(table map[string]NavInfo, parent string, items []SidebarItem) error {
	type entry struct {
		url  string
		href string
	}

	seen := make(map[string]bool)
	entries := make([]entry, 0)

	var walk func([]SidebarItem) error
	walk = func(items []SidebarItem) error {
		for _, item := range items {
			if item.Link != "" {
				_, target, external := app.resolveConfigLink(item.Link)
				if !external && target != "" {
					if page := app.findPageByURL(target); page != nil && !seen[page.URL] {
						if !routeContains(parent, page.URL) {
							return fmt.Errorf("sidebar %s links to %s outside its route", parent, page.SourcePath)
						}

						seen[page.URL] = true
						entries = append(entries, entry{
							url:  page.URL,
							href: page.Href,
						})
					}
				}
			}

			if err := walk(item.Items); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(items); err != nil {
		return err
	}

	for index, entry := range entries {
		info := NavInfo{}
		if index > 0 {
			info.PrevHref = entries[index-1].href
		}
		if index < len(entries)-1 {
			info.NextHref = entries[index+1].href
		}

		if _, ok := table[entry.url]; ok {
			return fmt.Errorf("page %s appears in more than one sidebar route", entry.url)
		}
		table[entry.url] = info
	}

	return nil
}

func routeContains(parent string, pageURL string) bool {
	if parent == "" || parent == "/" {
		return true
	}

	return sidebarPrefixMatches(parent, pageURL) || strings.TrimSuffix(parent, "/") == pageURL
}

func (app *App) lookupURLInfo(pageURL string) (NavInfo, bool) {
	info, ok := app.URLTab[pageURL]

	return info, ok
}

func (app *App) applyNavigation(state *RenderState, page *Page) {
	info, ok := app.lookupURLInfo(page.URL)
	if !ok {
		return
	}

	if info.PrevHref != "" {
		state.Prev = &PageLink{Text: app.titleForHref(info.PrevHref), Href: info.PrevHref}
	}
	if info.NextHref != "" {
		state.Next = &PageLink{Text: app.titleForHref(info.NextHref), Href: info.NextHref}
	}
}
