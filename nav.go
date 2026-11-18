package main

import "strings"

type NavLink struct {
	Text     string
	Href     string
	Path     string
	Active   bool
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
		link := &NavLink{
			Text:     strings.TrimSpace(item.Text),
			Href:     item.Link,
		}

		if link.Text == "" {
			link.Text = item.Link
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
		if pageURL == link.Href {
			link.Active = true
		}
		links = append(links, link)
	}

	return links
}

func (app *App) resolveSidebarTrees() map[string][]*SidebarNode {
	trees := make(map[string][]*SidebarNode)

	if app.Sidebar == nil {
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
		Href: item.Link,
		Text:      strings.TrimSpace(item.Text),
		Collapsed: item.Collapsed,
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

	nodes := app.matchedSidebar(pageURL)

	return app.markSidebar(nodes, pageURL)
}

func (app *App) markSidebar(nodes []*SidebarNode, pageURL string) []*SidebarNode {
	if len(nodes) == 0 {
		return nil
	}

	marked := make([]*SidebarNode, 0, len(nodes))
	for _, node := range nodes {
		clone := *node

		if len(clone.Items) != 0 {
			clone.Items = app.markSidebar(node.Items, pageURL)
		}

		clone.Active = node.Path != "" && node.Path == pageURL
		if !clone.Active && clone.Path != "" && len(clone.Items) > 0 {
			clone.Active = strings.HasPrefix(pageURL, strings.TrimSuffix(clone.Path, "/")+"/")
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
