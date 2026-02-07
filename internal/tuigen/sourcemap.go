package tuigen

import (
	"encoding/json"
	"strings"
)

// SourceMap tracks position mappings between .gsx source and generated .go files.
// All line numbers are 0-indexed.
type SourceMap struct {
	// SourceFile is the original .gsx file path
	SourceFile string `json:"sourceFile"`

	// Mappings contains position mappings from .go to .gsx
	Mappings []SourceMapping `json:"mappings"`
}

// SourceMapping represents a single line/column mapping.
type SourceMapping struct {
	// GoLine is the line in the generated .go file (0-indexed)
	GoLine int `json:"goLine"`
	// GoCol is the column in the generated .go file (0-indexed)
	GoCol int `json:"goCol"`
	// GsxLine is the line in the source .gsx file (0-indexed)
	GsxLine int `json:"gsxLine"`
	// GsxCol is the column in the source .gsx file (0-indexed)
	GsxCol int `json:"gsxCol"`
	// Length is the length of the mapped region
	Length int `json:"length"`
}

// NewSourceMap creates a new empty source map.
func NewSourceMap(sourceFile string) *SourceMap {
	return &SourceMap{
		SourceFile: sourceFile,
		Mappings:   make([]SourceMapping, 0),
	}
}

// AddMapping adds a new position mapping.
func (sm *SourceMap) AddMapping(m SourceMapping) {
	sm.Mappings = append(sm.Mappings, m)
}

// GoToGsx converts a .go position to a .gsx position.
// Returns the translated position and true if found, otherwise returns
// the input position and false.
func (sm *SourceMap) GoToGsx(goLine, goCol int) (gsxLine, gsxCol int, found bool) {
	for _, m := range sm.Mappings {
		if m.GoLine == goLine && goCol >= m.GoCol && goCol <= m.GoCol+m.Length {
			offset := goCol - m.GoCol
			return m.GsxLine, m.GsxCol + offset, true
		}
	}
	return goLine, goCol, false
}

// GsxToGo converts a .gsx position to a .go position.
// Returns the translated position and true if found.
func (sm *SourceMap) GsxToGo(gsxLine, gsxCol int) (goLine, goCol int, found bool) {
	for _, m := range sm.Mappings {
		if m.GsxLine == gsxLine && gsxCol >= m.GsxCol && gsxCol <= m.GsxCol+m.Length {
			offset := gsxCol - m.GsxCol
			return m.GoLine, m.GoCol + offset, true
		}
	}
	return gsxLine, gsxCol, false
}

// ToJSON serializes the source map to JSON.
func (sm *SourceMap) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sm, "", "  ")
}

// ParseSourceMap parses a source map from JSON.
func ParseSourceMap(data []byte) (*SourceMap, error) {
	var sm SourceMap
	if err := json.Unmarshal(data, &sm); err != nil {
		return nil, err
	}
	return &sm, nil
}

// SourceMapFileName returns the source map filename for a given generated file.
// e.g., "counter_gsx.go" -> "counter_gsx.go.map"
func SourceMapFileName(goFile string) string {
	return goFile + ".map"
}

// countLines counts the number of lines in a string.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}
