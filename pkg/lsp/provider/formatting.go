package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/pkg/formatter"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
)

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

// formattingProvider implements FormattingProvider.
type formattingProvider struct{}

// NewFormattingProvider creates a new formatting provider.
func NewFormattingProvider() FormattingProvider {
	return &formattingProvider{}
}

func (f *formattingProvider) Format(doc *Document, opts FormattingOptions) ([]TextEdit, error) {
	log.Server("Formatting provider for %s", doc.URI)

	// Create formatter with options
	fmtr := formatter.New()
	if opts.InsertSpaces {
		fmtr.IndentString = strings.Repeat(" ", opts.TabSize)
	} else {
		fmtr.IndentString = "\t"
	}

	// Format the document
	formatted, err := fmtr.Format(doc.URI, doc.Content)
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
