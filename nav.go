package main

import "strings"

type NavLink struct {
	Text        string
	Href        string
	ActiveMatch string
	Active      bool
}

func (app *App) resolveNav() []*NavLink {
	if len(app.Config.Nav) == 0 {
		return nil
	}

	links := make([]*NavLink, 0, len(app.Config.Nav))

	for _, item := range app.Config.Nav {
		link := &NavLink{
			Text:        strings.TrimSpace(item.Text),
			Href:        app.fmtURL(item.Link),
			ActiveMatch: item.ActiveMatch,
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
		clone.Active = clone.Href == pageURL || (clone.ActiveMatch != "" && strings.HasPrefix(pageURL, clone.ActiveMatch))
		links = append(links, &clone)
	}

	return links
}
