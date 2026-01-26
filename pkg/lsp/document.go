package lsp

import (
	"sync"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// Document represents an open .tui file with its parsed state.
type Document struct {
	URI     string
	Content string
	Version int
	AST     *tuigen.File
	Errors  []*tuigen.Error
}

// DocumentManager tracks all open documents.
type DocumentManager struct {
	mu   sync.RWMutex
	docs map[string]*Document
}

// NewDocumentManager creates a new document manager.
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		docs: make(map[string]*Document),
	}
}

// Open opens a new document and parses it.
func (dm *DocumentManager) Open(uri, content string, version int) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := &Document{
		URI:     uri,
		Content: content,
		Version: version,
	}

	dm.parseDocument(doc)
	dm.docs[uri] = doc
	return doc
}

// Update updates an existing document with new content.
func (dm *DocumentManager) Update(uri, content string, version int) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc, ok := dm.docs[uri]
	if !ok {
		// Document wasn't open, open it
		doc = &Document{
			URI:     uri,
			Content: content,
			Version: version,
		}
		dm.docs[uri] = doc
	} else {
		doc.Content = content
		doc.Version = version
	}

	dm.parseDocument(doc)
	return doc
}

// Close closes a document.
func (dm *DocumentManager) Close(uri string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.docs, uri)
}

// Get retrieves a document by URI.
func (dm *DocumentManager) Get(uri string) *Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.docs[uri]
}

// All returns all open documents.
func (dm *DocumentManager) All() []*Document {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	docs := make([]*Document, 0, len(dm.docs))
	for _, doc := range dm.docs {
		docs = append(docs, doc)
	}
	return docs
}

// parseDocument parses the document content and updates AST/Errors.
func (dm *DocumentManager) parseDocument(doc *Document) {
	// Extract filename from URI for error reporting
	filename := uriToPath(doc.URI)

	lexer := tuigen.NewLexer(filename, doc.Content)
	parser := tuigen.NewParser(lexer)
	ast, err := parser.ParseFile()

	doc.AST = ast

	// Collect errors
	doc.Errors = nil
	if err != nil {
		if errList, ok := err.(*tuigen.ErrorList); ok {
			doc.Errors = errList.Errors()
		} else if tuiErr, ok := err.(*tuigen.Error); ok {
			doc.Errors = []*tuigen.Error{tuiErr}
		}
	}

	// Run analyzer to collect semantic errors (including Tailwind class validation)
	if ast != nil {
		analyzer := tuigen.NewAnalyzer()
		if analyzerErr := analyzer.Analyze(ast); analyzerErr != nil {
			if errList, ok := analyzerErr.(*tuigen.ErrorList); ok {
				doc.Errors = append(doc.Errors, errList.Errors()...)
			} else if tuiErr, ok := analyzerErr.(*tuigen.Error); ok {
				doc.Errors = append(doc.Errors, tuiErr)
			}
		}
	}
}

// uriToPath converts a file:// URI to a file path.
func uriToPath(uri string) string {
	// Simple conversion - strip file:// prefix
	const prefix = "file://"
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		return uri[len(prefix):]
	}
	return uri
}

// Position represents a position in a document (0-indexed).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// PositionToOffset converts a Position to a byte offset in the content.
func PositionToOffset(content string, pos Position) int {
	line := 0
	offset := 0

	for i, ch := range content {
		if line == pos.Line {
			// Found the line, now count characters
			charCount := 0
			for j := i; j < len(content); j++ {
				if charCount == pos.Character {
					return j
				}
				if content[j] == '\n' {
					break
				}
				charCount++
			}
			return i + pos.Character
		}
		if ch == '\n' {
			line++
		}
		offset = i + 1
	}

	return offset
}

// OffsetToPosition converts a byte offset to a Position.
func OffsetToPosition(content string, offset int) Position {
	line := 0
	col := 0

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return Position{Line: line, Character: col}
}

// TuigenPosToRange converts a tuigen.Position to an LSP Range.
// tuigen positions are 1-indexed, LSP positions are 0-indexed.
func TuigenPosToRange(pos tuigen.Position, length int) Range {
	start := Position{
		Line:      pos.Line - 1,
		Character: pos.Column - 1,
	}
	end := Position{
		Line:      pos.Line - 1,
		Character: pos.Column - 1 + length,
	}
	return Range{Start: start, End: end}
}

// TuigenPosToRangeWithEnd converts start and end tuigen.Positions to an LSP Range.
// tuigen positions are 1-indexed, LSP positions are 0-indexed.
func TuigenPosToRangeWithEnd(startPos, endPos tuigen.Position) Range {
	start := Position{
		Line:      startPos.Line - 1,
		Character: startPos.Column - 1,
	}
	end := Position{
		Line:      endPos.Line - 1,
		Character: endPos.Column - 1,
	}
	return Range{Start: start, End: end}
}
