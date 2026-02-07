package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// SourceMapCache caches source maps loaded from disk.
type SourceMapCache struct {
	mu    sync.RWMutex
	maps  map[string]*tuigen.SourceMap // gsxURI -> source map
}

// NewSourceMapCache creates a new source map cache.
func NewSourceMapCache() *SourceMapCache {
	return &SourceMapCache{
		maps: make(map[string]*tuigen.SourceMap),
	}
}

// Get retrieves a source map for a .gsx file, loading from disk if needed.
func (c *SourceMapCache) Get(gsxURI string) *tuigen.SourceMap {
	c.mu.RLock()
	sm, ok := c.maps[gsxURI]
	c.mu.RUnlock()
	if ok {
		return sm
	}

	// Try to load from disk
	sm = c.loadFromDisk(gsxURI)
	if sm != nil {
		c.mu.Lock()
		c.maps[gsxURI] = sm
		c.mu.Unlock()
	}
	return sm
}

// Invalidate removes a cached source map, forcing reload on next access.
func (c *SourceMapCache) Invalidate(gsxURI string) {
	c.mu.Lock()
	delete(c.maps, gsxURI)
	c.mu.Unlock()
}

// loadFromDisk attempts to load a source map for the given .gsx URI.
func (c *SourceMapCache) loadFromDisk(gsxURI string) *tuigen.SourceMap {
	// Convert URI to file path
	gsxPath := uriToPath(gsxURI)
	if gsxPath == "" {
		return nil
	}

	// Compute the source map path: counter.gsx -> counter_gsx.go.map
	dir := filepath.Dir(gsxPath)
	base := filepath.Base(gsxPath)
	name := strings.TrimSuffix(base, ".gsx")
	name = strings.ReplaceAll(name, "-", "_")
	sourceMapPath := filepath.Join(dir, name+"_gsx.go.map")

	log.Server("Loading source map from %s", sourceMapPath)

	data, err := os.ReadFile(sourceMapPath)
	if err != nil {
		log.Server("Failed to read source map %s: %v", sourceMapPath, err)
		return nil
	}

	sm, err := tuigen.ParseSourceMap(data)
	if err != nil {
		log.Server("Failed to parse source map %s: %v", sourceMapPath, err)
		return nil
	}

	log.Server("Loaded source map with %d mappings", len(sm.Mappings))
	return sm
}

// Note: uriToPath is defined in document.go
