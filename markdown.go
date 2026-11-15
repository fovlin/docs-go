package main

import "acovia.net/docs-go/internal/markdown"

// OutlineEntry is the parser heading shape used by the layout.
type OutlineEntry = markdown.Heading

func parseMarkdown(data []byte) *markdown.Document {
	return markdown.Parse(data)
}

func renderMarkdown(document *markdown.Document, resolve markdown.ResolveFunc) []byte {
	return document.Render(resolve)
}
