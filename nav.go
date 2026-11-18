package main

import "strings"

type NavLink struct {
	Text   string
	Href   string
	Active bool
}

func (app *App) resolveNav() []*NavLink {
	if len(app.Config.Nav) == 0 {
		return nil
	}

	links := make([]*NavLink, 0, len(app.Config.Nav))

	for _, item := range app.Config.Nav {
		link := &NavLink{
			Text: strings.TrimSpace(item.Text),
			Href: item.Link,
		}

		if link.Text == "" {
			link.Text = item.Link
		}

		links = append(links, link)
	}

	return links
}

func (app *App) navFor(pageURL string) []*NavLink {
	if len(app.Nav) == 0 {
		return nil
	}

	links := make([]*NavLink, 0, len(app.Nav))
	for _, link := range app.Nav {
		clone := *link
		clone.Active = clone.Href == pageURL
		links = append(links, &clone)
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
		Href:      item.Link,
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
