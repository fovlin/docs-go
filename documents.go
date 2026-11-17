package main

import (
	"sync"

	"acovia.net/docs-go/internal/markdown"
)

type documentEntry struct {
	document *markdown.Document
	mod      int64
	size     int64
}

// DocumentCache maps absolute source paths to parsed documents. Entries are
// validated with modtime and size, so an edited file is parsed again on the
// next lookup. Parsed documents are immutable, so sharing them is safe.
type DocumentCache struct {
	mu      sync.RWMutex
	entries map[string]documentEntry
}

func newDocumentCache() *DocumentCache {
	return &DocumentCache{entries: make(map[string]documentEntry)}
}

func (cache *DocumentCache) lookup(absPath string, mod int64, size int64) (*markdown.Document, bool) {
	if cache == nil {
		return nil, false
	}

	cache.mu.RLock()
	entry, ok := cache.entries[absPath]
	cache.mu.RUnlock()

	if !ok || entry.mod != mod || entry.size != size {
		return nil, false
	}

	return entry.document, true
}

func (cache *DocumentCache) store(absPath string, document *markdown.Document, mod int64, size int64) {
	if cache == nil {
		return
	}

	cache.mu.Lock()
	cache.entries[absPath] = documentEntry{document: document, mod: mod, size: size}
	cache.mu.Unlock()
}
