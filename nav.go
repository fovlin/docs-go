package main

import "strings"

type NavLink struct {
	Text     string
	Href     string
	Path     string
	External bool
	Active   bool
	// ActiveMatch is the normalized URL prefix from config. It is resolved
	// once at load time; navFor only evaluates it against one page URL.
	ActiveMatch string
}

// resolveNav resolves configured navbar links once so rendering only has to
// mark the active entry.
func (app *App) resolveNav() []*NavLink {
	if len(app.Config.Nav) == 0 {
		return nil
	}

	links := make([]*NavLink, 0, len(app.Config.Nav))

	for _, item := range app.Config.Nav {
		linkHref, sitePath, external := app.resolveConfigLink(item.Link)
		link := &NavLink{
			Text:     strings.TrimSpace(item.Text),
			Href:     linkHref,
			Path:     sitePath,
			External: external,
		}

		if link.Text == "" && sitePath == "" {
			link.Text = item.Link
		}

		if !external && sitePath != "" && strings.TrimSpace(item.ActiveMatch) != "" {
			link.ActiveMatch = normalizeSidebarPrefix(item.ActiveMatch, app.Base)
		}

		links = append(links, link)
	}

	return links
}

// navFor returns a render-local copy of the resolved navbar with the active
// entry marked for pageURL.
func (app *App) navFor(pageURL string) []*NavLink {
	if len(app.Nav) == 0 {
		return nil
	}

	links := make([]*NavLink, 0, len(app.Nav))
	for _, link := range app.Nav {
		clone := *link
		if clone.Text == "" && clone.Path != "" {
			clone.Text = app.pageTitle(clone.Path)
		}
		if !clone.External && clone.Path != "" {
			if clone.ActiveMatch != "" {
				clone.Active = sidebarPrefixMatches(clone.ActiveMatch, pageURL)
			} else {
				clone.Active = navActive(clone.Path, pageURL)
			}
		}

		links = append(links, &clone)
	}

	return links
}

func navActive(sitePath string, pageURL string) bool {
	if sitePath == "" {
		return false
	}

	if sitePath == "/" {
		return pageURL == "/"
	}

	if sitePath == pageURL {
		return true
	}

	return strings.HasPrefix(pageURL, strings.TrimSuffix(sitePath, "/")+"/")
}

// resolveSidebarTrees resolves the sidebar link text, href, path and external
// flag for every configured route once. Active/open state is per page.
func (app *App) resolveSidebarTrees() map[string][]*SidebarNode {
	trees := make(map[string][]*SidebarNode)

	if app.Sidebar == nil {
		return trees
	}

	if app.Sidebar.Global != nil {
		trees["/"] = app.resolveSidebarItems(app.Sidebar.Global)
		return trees
	}

	for _, key := range app.Sidebar.keys {
		trees[key] = app.resolveSidebarItems(app.Sidebar.Sections[key])
	}

	return trees
}

func (app *App) resolveSidebarItems(items []SidebarItem) []*SidebarNode {
	nodes := make([]*SidebarNode, 0, len(items))
	for index := range items {
		if node := app.resolveSidebarNode(items[index]); node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func (app *App) resolveSidebarNode(item SidebarItem) *SidebarNode {
	node := &SidebarNode{
		Text:      strings.TrimSpace(item.Text),
		Collapsed: item.Collapsed,
	}

	if item.Link != "" {
		node.Href, node.Path, node.External = app.resolveConfigLink(item.Link)
	}

	for index := range item.Items {
		if child := app.resolveSidebarNode(item.Items[index]); child != nil {
			node.Items = append(node.Items, child)
		}
	}

	if node.Href == "" && len(node.Items) == 0 {
		return nil
	}

	return node
}

// sidebarFor returns a render-local copy of the matching resolved tree with
// active/open state marked for pageURL.
func (app *App) sidebarFor(pageURL string) []*SidebarNode {
	if app.Sidebar == nil {
		return nil
	}

	var nodes []*SidebarNode
	if app.Sidebar.Global != nil {
		nodes = app.SidebarTrees["/"]
	} else {
		for _, key := range app.Sidebar.keys {
			if sidebarPrefixMatches(key, pageURL) {
				nodes = app.SidebarTrees[key]
				break
			}
		}
	}

	return app.markSidebar(nodes, pageURL)
}

func (app *App) markSidebar(nodes []*SidebarNode, pageURL string) []*SidebarNode {
	if len(nodes) == 0 {
		return nil
	}

	marked := make([]*SidebarNode, 0, len(nodes))
	for _, node := range nodes {
		clone := *node
		clone.Items = app.markSidebar(node.Items, pageURL)

		clone.Active = node.Path != "" && node.Path == pageURL
		if !clone.Active && clone.Path != "" && len(clone.Items) > 0 {
			clone.Active = strings.HasPrefix(pageURL, strings.TrimSuffix(clone.Path, "/")+"/")
		}

		if clone.Text == "" {
			if clone.Path != "" {
				clone.Text = app.pageTitle(clone.Path)
			} else if len(clone.Items) > 0 {
				clone.Text = clone.Items[0].Text
				if clone.Text == "" {
					clone.Text = "Section"
				}
			}
		}

		clone.HasActive = false
		for _, child := range clone.Items {
			if child.Active || child.HasActive {
				clone.HasActive = true
				break
			}
		}

		clone.Open = !clone.Collapsed || clone.Active || clone.HasActive
		marked = append(marked, &clone)
	}

	return marked
}
