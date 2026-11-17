package main

import "strings"

// pageTitle resolves the display title for a configured site path. It is used
// at load time to pre-resolve navbar and sidebar labels.
func (app *App) pageTitle(sitePath string) string {
	if page := app.findPageByURL(sitePath); page != nil {
		return app.titleFor(page)
	}

	if page := app.findPageBySource(strings.TrimPrefix(sitePath, "/")); page != nil {
		return app.titleFor(page)
	}

	return humanizeTitle(strings.TrimSuffix(sitePath, "/"))
}

func (app *App) titleFor(page *Page) string {
	document, err := app.parsePage(page)
	if err == nil && document.FirstH1 != "" {
		return document.FirstH1
	}

	return humanizeTitle(page.SourcePath)
}
