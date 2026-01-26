package lsp

import (
	"encoding/json"
	"strings"

	"github.com/grindlemire/go-tui/pkg/formatter"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
)

// DocumentFormattingParams represents textDocument/formatting parameters.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions represents formatting options.
type FormattingOptions struct {
	TabSize                int  `json:"tabSize"`
	InsertSpaces           bool `json:"insertSpaces"`
	TrimTrailingWhitespace bool `json:"trimTrailingWhitespace,omitempty"`
	InsertFinalNewline     bool `json:"insertFinalNewline,omitempty"`
	TrimFinalNewlines      bool `json:"trimFinalNewlines,omitempty"`
}

// TextEdit represents a text edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// handleFormatting handles textDocument/formatting requests.
func (s *Server) handleFormatting(params json.RawMessage) (any, *Error) {
	var p DocumentFormattingParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Formatting request for %s", p.TextDocument.URI)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}

	// Create formatter with options
	f := formatter.New()
	if p.Options.InsertSpaces {
		f.IndentString = strings.Repeat(" ", p.Options.TabSize)
	} else {
		f.IndentString = "\t"
	}

	// Format the document
	formatted, err := f.Format(p.TextDocument.URI, doc.Content)
	if err != nil {
		log.Server("Formatting error: %v", err)
		// Return empty edits on error - don't fail the request
		return []TextEdit{}, nil
	}

	// If nothing changed, return empty edits
	if formatted == doc.Content {
		return []TextEdit{}, nil
	}

	// Calculate the range of the entire document
	lines := strings.Split(doc.Content, "\n")
	lastLine := len(lines) - 1
	lastChar := 0
	if lastLine >= 0 && len(lines[lastLine]) > 0 {
		lastChar = len(lines[lastLine])
	}

	// Return a single edit that replaces the entire document
	edits := []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}

	return edits, nil
}
