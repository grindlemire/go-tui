// Package formatter provides code formatting for .tui files.
package formatter

import (
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// Formatter formats .tui source code.
type Formatter struct {
	// IndentString is the string used for indentation (default: tab).
	IndentString string
	// MaxLineWidth is the target maximum line width (default: 100).
	MaxLineWidth int
}

// New creates a new Formatter with default settings.
func New() *Formatter {
	return &Formatter{
		IndentString: "\t",
		MaxLineWidth: 100,
	}
}

// Format parses and reformats the given .tui source code.
// Returns the formatted code and any error encountered during parsing.
func (f *Formatter) Format(filename, source string) (string, error) {
	// Parse the source
	lexer := tuigen.NewLexer(filename, source)
	parser := tuigen.NewParser(lexer)

	file, err := parser.ParseFile()
	if err != nil {
		return "", err
	}

	// Generate formatted output
	printer := newPrinter(f.IndentString, f.MaxLineWidth)
	return printer.PrintFile(file), nil
}

// FormatResult contains the result of formatting a file.
type FormatResult struct {
	// Content is the formatted content.
	Content string
	// Changed indicates if the content was different from the original.
	Changed bool
}

// FormatWithResult formats the source and indicates if it changed.
func (f *Formatter) FormatWithResult(filename, source string) (FormatResult, error) {
	formatted, err := f.Format(filename, source)
	if err != nil {
		return FormatResult{}, err
	}

	return FormatResult{
		Content: formatted,
		Changed: formatted != source,
	}, nil
}
