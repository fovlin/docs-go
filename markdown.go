package main

import "acovia.net/docs-go/internal/markdown"

// OutlineEntry is the parser heading shape used by the layout.
type OutlineEntry = markdown.Heading

func parseMarkdown(data []byte, config Config) (*markdown.Document, error) {
	return markdown.Parse(data, markdown.Options{
		Containers: config.Containers,
	}), nil
}

func renderMarkdown(document *markdown.Document, resolve markdown.ResolveFunc) []byte {
	return document.Render(resolve)
}
