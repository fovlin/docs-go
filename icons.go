package main

import "strings"

var socialPaths = map[string]string{
	"github":   `<path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.7c-2.78.6-3.37-1.34-3.37-1.34-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.89 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.56-1.11-4.56-4.94 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02a9.5 9.5 0 0 1 5 0c1.91-1.29 2.75-1.02 2.75-1.02.55 1.38.2 2.4.1 2.65.64.7 1.03 1.59 1.03 2.68 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.85v2.74c0 .27.18.58.69.48A10 10 0 0 0 12 2Z"/>`,
	"gitlab":   `<path d="m12 21 3.7-11.4H8.3L12 21Z"/><path d="M3.3 9.6 2 13.9c-.1.4 0 .8.3 1l9.7 7.1-8.7-12.4Z"/><path d="M3.3 9.6h5L6.2 3.1c-.1-.3-.5-.3-.6 0L3.3 9.6Z"/><path d="m20.7 9.6 1.3 4.3c.1.4 0 .8-.3 1L12 21.9l8.7-12.3Z"/><path d="M20.7 9.6h-5l2.1-6.5c.1-.3.5-.3.6 0l2.3 6.5Z"/>`,
	"x":        `<path d="M4 3h4.2l5 6.6L18.8 3H21l-6.6 7.6L21.5 21h-4.2l-5.4-7.1L6 21H3.8l7-8.1L4 3Z"/>`,
	"twitter":  `<path d="M4 3h4.2l5 6.6L18.8 3H21l-6.6 7.6L21.5 21h-4.2l-5.4-7.1L6 21H3.8l7-8.1L4 3Z"/>`,
	"discord":  `<path d="M8.5 5.5A15 15 0 0 0 5 7.3C3 11 2.4 14.4 2.7 17.4a15 15 0 0 0 4.6 2.3l.9-1.5c-.5-.2-1-.5-1.5-.8l.4-.3a10.7 10.7 0 0 0 9.8 0l.4.3c-.5.3-1 .6-1.5.8l.9 1.5a15 15 0 0 0 4.6-2.3c.4-3.5-.6-7-2.3-10.1a15 15 0 0 0-3.5-1.8l-.8 1.6a12 12 0 0 0-3.4 0L11 3.9l-.8.4A15 15 0 0 0 8.5 5.5Z"/><circle cx="9" cy="13" r="1.3" fill="currentColor" stroke="none"/><circle cx="15" cy="13" r="1.3" fill="currentColor" stroke="none"/>`,
	"youtube":  `<rect x="2.5" y="5" width="19" height="14" rx="4"/><path d="m10 9 5 3-5 3V9Z"/>`,
	"linkedin": `<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M8 10v7M8 7v.01M12 17v-4a2 2 0 0 1 4 0v4M12 10v7"/>`,
	"rss":      `<path d="M4 11a9 9 0 0 1 9 9M4 4a16 16 0 0 1 16 16"/><circle cx="5" cy="19" r="1.5" fill="currentColor" stroke="none"/>`,
	"mail":     `<rect x="3" y="5" width="18" height="14" rx="2"/><path d="m3 7 9 6 9-6"/>`,
	"email":    `<rect x="3" y="5" width="18" height="14" rx="2"/><path d="m3 7 9 6 9-6"/>`,
	"link":     `<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18"/>`,
}

var filledSocialIcons = map[string]bool{
	"github":  true,
	"gitlab":  true,
	"x":       true,
	"twitter": true,
	"discord": true,
}

func socialIcon(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	path, ok := socialPaths[key]
	if !ok {
		key = "link"
		path = socialPaths[key]
	}

	attributes := `fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"`
	if filledSocialIcons[key] {
		attributes = `fill="currentColor" stroke="none"`
	}

	return `<svg viewBox="0 0 24 24" ` + attributes + ` aria-hidden="true">` + path + `</svg>`
}
