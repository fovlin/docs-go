package main

import (
	"acovia.net/docs-go/internal/markdown"
)

type OutlineEntry struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id"`
}

type parsedMarkdown struct {
	Document       *markdown.Document
	Headings       []OutlineEntry
	FirstH1        string
	FirstParagraph string
	PlainText      string
}

func parseMarkdown(data []byte, config Config) (*parsedMarkdown, error) {
	document := markdown.Parse(data, markdown.Options{
		Containers: config.Containers,
	})

	headings := make([]OutlineEntry, 0, len(document.Headings))
	for _, heading := range document.Headings {
		headings = append(headings, OutlineEntry{
			Level: heading.Level,
			Text:  heading.Text,
			ID:    heading.ID,
		})
	}

	return &parsedMarkdown{
		Document:       document,
		Headings:       headings,
		FirstH1:        document.FirstH1,
		FirstParagraph: document.FirstParagraph,
		PlainText:      document.PlainText,
	}, nil
}

func renderMarkdown(parsed *parsedMarkdown, resolve markdown.ResolveFunc) []byte {
	return parsed.Document.Render(resolve)
}
