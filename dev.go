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
	"sync/atomic"
	"time"

	"acovia.net/docs-go/internal/record"
)

type Dev struct {
	app      atomic.Pointer[App]
	reloadMu sync.Mutex
	stamps   atomic.Pointer[devStamps]

	root       string
	configFile string
	srcFlag    string
	outFlag    string
}

type devStamps struct {
	config  string
	sidebar string
	src     string
}

func runDev(app *App, addr string, port int) error {
	return newDev(app, configFile, srcFlag, outFlag).run(addr, port)
}

func newDev(app *App, configFile, srcFlag, outFlag string) *Dev {
	dev := &Dev{
		configFile: configFile,
		srcFlag:    srcFlag,
		outFlag:    outFlag,
	}
	dev.app.Store(app)
	dev.stamps.Store(&devStamps{
		config:  fileStamp(app.ConfigPath),
		sidebar: fileStamp(app.SidebarPath),
		src:     fileStamp(app.SrcDir),
	})

	return dev
}

func (dev *Dev) run(addr string, port int) error {
	server := &http.Server{
		Addr:              net.JoinHostPort(addr, strconv.Itoa(port)),
		Handler:           http.HandlerFunc(dev.handle),
		ReadHeaderTimeout: 10 * time.Second,
	}

	record.Info(fmt.Sprintf("dev server: http://%s", server.Addr))

	return server.ListenAndServe()
}

func (dev *Dev) handle(w http.ResponseWriter, r *http.Request) {
	app := dev.snapshot()
	if app == nil {
		http.Error(w, "no project snapshot", http.StatusInternalServerError)
		return
	}

	app.serveDev(w, r)
}

// snapshot returns the current App. It reloads when the config or sidebar file
// changed and keeps the previous snapshot if the reload fails.
func (dev *Dev) snapshot() *App {
	app := dev.app.Load()
	if app == nil {
		return nil
	}

	stamps := dev.stamps.Load()
	configStamp := fileStamp(app.ConfigPath)
	sidebarStamp := fileStamp(app.SidebarPath)
	srcStamp := fileStamp(app.SrcDir)
	if stamps != nil && configStamp == stamps.config && sidebarStamp == stamps.sidebar && srcStamp == stamps.src {
		return app
	}

	dev.reloadMu.Lock()
	defer dev.reloadMu.Unlock()

	app = dev.app.Load()
	if app == nil {
		return nil
	}

	stamps = dev.stamps.Load()
	configStamp = fileStamp(app.ConfigPath)
	sidebarStamp = fileStamp(app.SidebarPath)
	srcStamp = fileStamp(app.SrcDir)
	if stamps != nil && configStamp == stamps.config && sidebarStamp == stamps.sidebar && srcStamp == stamps.src {
		return app
	}

	reloaded, err := LoadApp(dev.root, dev.configFile, dev.srcFlag, dev.outFlag)
	if err != nil {
		record.Info(fmt.Sprintf("reload: %v", err))
		// Remember the failed stamps so an unchanged broken config is not
		// retried on every request; any later edit triggers a reload again.
		dev.stamps.Store(&devStamps{config: configStamp, sidebar: sidebarStamp, src: srcStamp})
		return app
	}

	dev.app.Store(reloaded)
	dev.stamps.Store(&devStamps{
		config:  fileStamp(reloaded.ConfigPath),
		sidebar: fileStamp(reloaded.SidebarPath),
		src:     fileStamp(reloaded.SrcDir),
	})

	return reloaded
}

func (app *App) serveDev(w http.ResponseWriter, r *http.Request) {
	sitePath := r.URL.Path
	if !strings.HasPrefix(sitePath, "/") {
		sitePath = "/" + sitePath
	}

	filePath, data, err := app.fileForURL(sitePath)
	switch {
	case err == nil && isMarkdownFile(filePath) && isWithin(app.SrcDir, filePath):
		page, err := app.loadPage(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		document, err := app.renderPageHTML(page, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		serveBytes(w, r, page.OutPath, document)
		return
	case err == nil:
		serveBytes(w, r, filepath.Base(filePath), data)
		return
	case !os.IsNotExist(err):
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch {
	case sitePath == "/favicon.ico":
		serveBytes(w, r, "favicon.ico", app.Favicon)
		return
	case strings.HasPrefix(sitePath, "/assets/"):
		serveThemeAsset(w, r, strings.TrimPrefix(sitePath, "/assets/"))
		return
	case sitePath == "/sitemap.xml" && app.Sitemap:
		document, err := app.sitemapBytes()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		serveBytes(w, r, "sitemap.xml", document)
		return
	}

	document := app.renderLayout(app.notFoundPage())
	if data, err := os.ReadFile(filepath.Join(app.PublicDir, "404.html")); err == nil {
		document = data
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	if r.Method != http.MethodHead {
		_, _ = w.Write(document)
	}
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

func serveBytes(w http.ResponseWriter, r *http.Request, name string, data []byte) {
	if contentType := mime.TypeByExtension(filepath.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}
