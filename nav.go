package main

import "strings"

type NavLink struct {
	Text     string
	Href     string
	Path     string
	External bool
	Active   bool
}

func buildNav(page *Page) []*NavLink {
	links := make([]*NavLink, 0, len(config.Nav))

	for _, item := range config.Nav {
		linkHref, sitePath, external := resolveConfigLink(item.Link)
		link := &NavLink{
			Text:     strings.TrimSpace(item.Text),
			Href:     linkHref,
			Path:     sitePath,
			External: external,
		}

		if link.Text == "" {
			if sitePath != "" {
				link.Text = pageTitle(sitePath)
			} else {
				link.Text = item.Link
			}
		}

		if !external && sitePath != "" {
			if strings.TrimSpace(item.ActiveMatch) != "" {
				link.Active = sidebarPrefixMatches(normalizeSidebarPrefix(item.ActiveMatch, config.Base), page.URL)
			} else {
				link.Active = navActive(sitePath, page.URL)
			}
		}

		links = append(links, link)
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

func buildSidebar(page *Page) []*SidebarNode {
	if sidebar == nil {
		return nil
	}

	items := sidebar.selectFor(page.URL)
	if len(items) == 0 {
		return nil
	}

	nodes := make([]*SidebarNode, 0, len(items))
	for index := range items {
		if node := sidebarNode(items[index], page.URL); node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func sidebarNode(item SidebarItem, pageURL string) *SidebarNode {
	node := &SidebarNode{
		Text:      strings.TrimSpace(item.Text),
		Collapsed: item.Collapsed,
	}

	if item.Link != "" {
		node.Href, node.Path, node.External = resolveConfigLink(item.Link)

		if node.Text == "" && node.Path != "" {
			node.Text = pageTitle(node.Path)
		}

		if !node.External && node.Path != "" {
			node.Active = node.Path == pageURL
		}
	}

	for index := range item.Items {
		if child := sidebarNode(item.Items[index], pageURL); child != nil {
			node.Items = append(node.Items, child)
		}
	}

	if node.Href == "" && len(node.Items) == 0 {
		return nil
	}

	if node.Text == "" && len(node.Items) > 0 {
		node.Text = node.Items[0].Text
		if node.Text == "" {
			node.Text = "Section"
		}
	}

	if !node.Active && node.Path != "" && len(node.Items) > 0 {
		node.Active = strings.HasPrefix(pageURL, strings.TrimSuffix(node.Path, "/")+"/")
	}

	for _, child := range node.Items {
		if child.Active || child.HasActive {
			node.HasActive = true
			break
		}
	}

	node.Open = !node.Collapsed || node.Active || node.HasActive

	return node
}
