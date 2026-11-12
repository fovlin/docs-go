package main

import "embed"

//go:embed theme scaffold
var themeFS embed.FS

const version = "1.0.0"
