package main

import (
	"net/url"
	"path"
	"strings"
)

func escapePath(sitePath string) string {
	if sitePath == "" {
		sitePath = "/"
	}

	if !strings.HasPrefix(sitePath, "/") {
		sitePath = "/" + sitePath
	}

	return (&url.URL{Path: sitePath}).EscapedPath()
}

func canonicalURL(target string) string {
	trailing := strings.HasSuffix(target, "/")
	cleaned := path.Clean("/" + target)
	if cleaned == "." {
		cleaned = "/"
	}

	if trailing && cleaned != "/" {
		cleaned += "/"
	}

	return cleaned
}