package gopls

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// mappingDebugLog is a package-level logger for debugging
var mappingDebugLog *os.File

// SetMappingDebugLog sets the debug log file for mapping operations
func SetMappingDebugLog(f *os.File) {
	mappingDebugLog = f
}

func logMappingDebug(format string, args ...any) {
	if mappingDebugLog != nil {
		fmt.Fprintf(mappingDebugLog, "[mapping] "+format+"\n", args...)
	}
}

// SourceMap tracks position mappings between .tui and generated .go files.
type SourceMap struct {
	mu       sync.RWMutex
	mappings []Mapping
}

// Mapping represents a position mapping between .tui and .go coordinates.
// All line and column values are 0-indexed.
type Mapping struct {
	// Position in .tui file
	TuiLine int
	TuiCol  int

	// Position in generated .go file
	GoLine int
	GoCol  int

	// Length of the mapped region
	Length int
}

// NewSourceMap creates a new empty source map.
func NewSourceMap() *SourceMap {
	return &SourceMap{
		mappings: make([]Mapping, 0),
	}
}

// AddMapping adds a new position mapping.
func (sm *SourceMap) AddMapping(m Mapping) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.mappings = append(sm.mappings, m)
}

// TuiToGo converts a .tui position to a .go position.
// Returns the translated position and true if a mapping was found,
// or the original position and false if no mapping covers this position.
func (sm *SourceMap) TuiToGo(tuiLine, tuiCol int) (goLine, goCol int, found bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	logMappingDebug("TuiToGo: looking for TuiLine=%d TuiCol=%d", tuiLine, tuiCol)
	for _, m := range sm.mappings {
		// Check if position is within this mapping (inclusive of start, exclusive of end)
		// OR if position is exactly at the exclusive end boundary (tuiCol == m.TuiCol+m.Length)
		// This handles LSP ranges which use exclusive end positions.
		if m.TuiLine == tuiLine && tuiCol >= m.TuiCol && tuiCol <= m.TuiCol+m.Length {
			// Position is within this mapping or at its exclusive end
			offset := tuiCol - m.TuiCol
			logMappingDebug("TuiToGo: MATCH found! mapping TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d -> result GoLine=%d GoCol=%d (offset=%d)",
				m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, m.GoLine, m.GoCol+offset, offset)
			return m.GoLine, m.GoCol + offset, true
		}
	}

	logMappingDebug("TuiToGo: NO MATCH found, returning original %d:%d", tuiLine, tuiCol)
	return tuiLine, tuiCol, false
}

// GoToTui converts a .go position to a .tui position.
// Returns the translated position and true if a mapping was found,
// or the original position and false if no mapping covers this position.
func (sm *SourceMap) GoToTui(goLine, goCol int) (tuiLine, tuiCol int, found bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	logMappingDebug("GoToTui: looking for GoLine=%d GoCol=%d", goLine, goCol)
	for _, m := range sm.mappings {
		// Check if position is within this mapping (inclusive of start, exclusive of end)
		// OR if position is exactly at the exclusive end boundary (goCol == m.GoCol+m.Length)
		// This handles LSP ranges which use exclusive end positions.
		if m.GoLine == goLine && goCol >= m.GoCol && goCol <= m.GoCol+m.Length {
			// Position is within this mapping or at its exclusive end
			offset := goCol - m.GoCol
			logMappingDebug("GoToTui: MATCH found! mapping GoLine=%d GoCol=%d TuiLine=%d TuiCol=%d Len=%d -> result TuiLine=%d TuiCol=%d (offset=%d)",
				m.GoLine, m.GoCol, m.TuiLine, m.TuiCol, m.Length, m.TuiLine, m.TuiCol+offset, offset)
			return m.TuiLine, m.TuiCol + offset, true
		}
	}

	logMappingDebug("GoToTui: NO MATCH found, returning original %d:%d", goLine, goCol)
	return goLine, goCol, false
}

// FindMappingForTuiPosition finds the mapping that contains the given .tui position.
// Returns nil if no mapping contains the position.
func (sm *SourceMap) FindMappingForTuiPosition(tuiLine, tuiCol int) *Mapping {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for i := range sm.mappings {
		m := &sm.mappings[i]
		// Include the exclusive end boundary for LSP compatibility
		if m.TuiLine == tuiLine && tuiCol >= m.TuiCol && tuiCol <= m.TuiCol+m.Length {
			return m
		}
	}

	return nil
}

// FindMappingForGoPosition finds the mapping that contains the given .go position.
// Returns nil if no mapping contains the position.
func (sm *SourceMap) FindMappingForGoPosition(goLine, goCol int) *Mapping {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for i := range sm.mappings {
		m := &sm.mappings[i]
		// Include the exclusive end boundary for LSP compatibility
		if m.GoLine == goLine && goCol >= m.GoCol && goCol <= m.GoCol+m.Length {
			return m
		}
	}

	return nil
}

// IsInGoExpression returns true if the given .tui position is within a Go expression.
func (sm *SourceMap) IsInGoExpression(tuiLine, tuiCol int) bool {
	return sm.FindMappingForTuiPosition(tuiLine, tuiCol) != nil
}

// AllMappings returns all mappings sorted by .tui position.
func (sm *SourceMap) AllMappings() []Mapping {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]Mapping, len(sm.mappings))
	copy(result, sm.mappings)

	sort.Slice(result, func(i, j int) bool {
		if result[i].TuiLine != result[j].TuiLine {
			return result[i].TuiLine < result[j].TuiLine
		}
		return result[i].TuiCol < result[j].TuiCol
	})

	return result
}

// Clear removes all mappings.
func (sm *SourceMap) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.mappings = sm.mappings[:0]
}

// Len returns the number of mappings.
func (sm *SourceMap) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.mappings)
}

// VirtualFileCache caches virtual .go files and their source maps.
type VirtualFileCache struct {
	mu    sync.RWMutex
	files map[string]*CachedVirtualFile
}

// CachedVirtualFile holds a cached virtual file and its metadata.
type CachedVirtualFile struct {
	TuiURI    string
	GoURI     string
	Content   string
	SourceMap *SourceMap
	Version   int
}

// NewVirtualFileCache creates a new virtual file cache.
func NewVirtualFileCache() *VirtualFileCache {
	return &VirtualFileCache{
		files: make(map[string]*CachedVirtualFile),
	}
}

// Get retrieves a cached virtual file by .tui URI.
func (c *VirtualFileCache) Get(tuiURI string) *CachedVirtualFile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.files[tuiURI]
}

// Put stores a virtual file in the cache.
func (c *VirtualFileCache) Put(tuiURI, goURI, content string, sourceMap *SourceMap, version int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[tuiURI] = &CachedVirtualFile{
		TuiURI:    tuiURI,
		GoURI:     goURI,
		Content:   content,
		SourceMap: sourceMap,
		Version:   version,
	}
}

// Remove removes a cached virtual file.
func (c *VirtualFileCache) Remove(tuiURI string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.files, tuiURI)
}

// GetByGoURI retrieves a cached virtual file by .go URI.
func (c *VirtualFileCache) GetByGoURI(goURI string) *CachedVirtualFile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, f := range c.files {
		if f.GoURI == goURI {
			return f
		}
	}
	return nil
}

// All returns all cached virtual files.
func (c *VirtualFileCache) All() []*CachedVirtualFile {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*CachedVirtualFile, 0, len(c.files))
	for _, f := range c.files {
		result = append(result, f)
	}
	return result
}
