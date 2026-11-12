package main

import (
	"bytes"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	devMu        sync.Mutex
	configStamp  string
	sidebarStamp string
)

func runDev(addr string, port int) error {
	configStamp = fileStamp(configPath)
	sidebarStamp = fileStamp(sidebarPath)

	server := &http.Server{
		Addr:              net.JoinHostPort(addr, strconv.Itoa(port)),
		Handler:           http.HandlerFunc(handleDev),
		ReadHeaderTimeout: 10 * time.Second,
	}

	logInfo("dev server: http://%s", server.Addr)

	return server.ListenAndServe()
}

func handleDev(w http.ResponseWriter, r *http.Request) {
	devMu.Lock()
	defer devMu.Unlock()

	refreshProject()

	sitePath := strings.TrimPrefix(r.URL.Path, config.Base)
	if sitePath == r.URL.Path && config.Base != "/" {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(sitePath, "/") {
		sitePath = "/" + sitePath
	}

	switch {
	case sitePath == "/favicon.ico":
		serveBytes(w, r, "favicon.ico", favicon)
		return
	case strings.HasPrefix(sitePath, "/assets/"):
		serveThemeAsset(w, r, strings.TrimPrefix(sitePath, "/assets/"))
		return
	case sitePath == "/sitemap.xml" && config.Sitemap:
		document, err := sitemapBytes()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		serveBytes(w, r, "sitemap.xml", document)
		return
	}

	if page := findPageByURL(sitePath); page != nil {
		if page.URL != "/" && strings.HasSuffix(page.URL, "/") && !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}

		document, err := renderPageHTML(page)
		if err != nil {
			logError("render %s: %v", page.SourcePath, err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		serveBytes(w, r, page.OutPath, document)
		return
	}

	if serveAsset(w, r, sitePath) {
		return
	}

	w.WriteHeader(http.StatusNotFound)
	serveBytes(w, r, "404.html", renderLayout(layout, notFoundPage()))
}

func refreshProject() {
	nextConfig := fileStamp(configPath)
	nextSidebar := fileStamp(sidebarPath)
	if nextConfig == configStamp && nextSidebar == sidebarStamp {
		return
	}

	if err := loadProject(root); err != nil {
		logWarn("reload: %v", err)
		return
	}

	configStamp = nextConfig
	sidebarStamp = nextSidebar
}

func fileStamp(name string) string {
	info, err := os.Stat(name)
	if err != nil {
		return "missing"
	}

	return fmt.Sprintf("%d:%d", info.ModTime().UnixNano(), info.Size())
}

func serveThemeAsset(w http.ResponseWriter, r *http.Request, name string) {
	if strings.Contains(name, "/") || name == "" {
		http.NotFound(w, r)
		return
	}

	data, err := themeFS.ReadFile("theme/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	serveBytes(w, r, name, data)
}

func serveAsset(w http.ResponseWriter, r *http.Request, sitePath string) bool {
	relative := strings.TrimPrefix(sitePath, "/")
	if relative == "" {
		return false
	}

	for _, directory := range []string{srcDir, publicDir} {
		target := filepath.Join(directory, filepath.FromSlash(relative))
		if !isInside(directory, target) {
			continue
		}

		info, err := os.Stat(target)
		if err != nil || info.IsDir() {
			continue
		}

		data, err := os.ReadFile(target)
		if err != nil {
			continue
		}

		serveBytes(w, r, relative, data)
		return true
	}

	return false
}

func serveBytes(w http.ResponseWriter, r *http.Request, name string, data []byte) {
	if contentType := mime.TypeByExtension(filepath.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}
