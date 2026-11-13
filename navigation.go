package main

// applyNavigation derives previous/next links from the sidebar that is
// currently active for the page. Different routes can therefore expose
// different navigation contexts for the same page.
func (app *App) applyNavigation(state *RenderState) {
	links := make([]*SidebarNode, 0)

	var walk func([]*SidebarNode)
	walk = func(nodes []*SidebarNode) {
		for _, node := range nodes {
			if node.Href != "" && node.Path != "" && !node.External {
				links = append(links, node)
			}
			walk(node.Items)
		}
	}
	walk(state.Sidebar)

	for index, node := range links {
		if node.Path != state.URL {
			continue
		}
		if index > 0 {
			prev := links[index-1]
			state.Prev = &PageLink{Text: prev.Text, Href: prev.Href}
		}
		if index < len(links)-1 {
			next := links[index+1]
			state.Next = &PageLink{Text: next.Text, Href: next.Href}
		}
		return
	}
}
