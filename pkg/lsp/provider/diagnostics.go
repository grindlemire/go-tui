package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
)

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	DiagnosticSeverityError         DiagnosticSeverity = 1
	DiagnosticSeverityWarning       DiagnosticSeverity = 2
	DiagnosticSeverityInformation   DiagnosticSeverity = 3
	DiagnosticSeverityHint          DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// diagnosticsProvider implements DiagnosticsProvider.
type diagnosticsProvider struct{}

// NewDiagnosticsProvider creates a new diagnostics provider.
func NewDiagnosticsProvider() DiagnosticsProvider {
	return &diagnosticsProvider{}
}

func (d *diagnosticsProvider) Diagnose(doc *Document) ([]Diagnostic, error) {
	log.Server("Diagnostics provider for %s", doc.URI)

	if doc.AST == nil && len(doc.Errors) == 0 {
		return []Diagnostic{}, nil
	}

	diagnostics := make([]Diagnostic, 0, len(doc.Errors))

	for _, err := range doc.Errors {
		var rng Range
		if err.EndPos.Line > 0 || err.EndPos.Column > 0 {
			rng = Range{
				Start: Position{
					Line:      err.Pos.Line - 1,
					Character: err.Pos.Column - 1,
				},
				End: Position{
					Line:      err.EndPos.Line - 1,
					Character: err.EndPos.Column - 1,
				},
			}
		} else {
			length := estimateErrorLength(err.Message)
			rng = Range{
				Start: Position{
					Line:      err.Pos.Line - 1,
					Character: err.Pos.Column - 1,
				},
				End: Position{
					Line:      err.Pos.Line - 1,
					Character: err.Pos.Column - 1 + length,
				},
			}
		}

		diag := Diagnostic{
			Range:    rng,
			Severity: DiagnosticSeverityError,
			Source:   "gsx",
			Message:  err.Message,
		}
		if err.Hint != "" {
			diag.Message = err.Message + " (" + err.Hint + ")"
		}
		diagnostics = append(diagnostics, diag)
	}

	return diagnostics, nil
}

// estimateErrorLength estimates the length of text to highlight for an error.
// It extracts quoted tokens from the message (e.g., "unexpected 'foo'") and uses
// their length, or falls back to the first word length if no quotes are found.
func estimateErrorLength(message string) int {
	// Try to extract a quoted token from the message (single or double quotes).
	for _, q := range []byte{'\'', '"', '`'} {
		start := strings.IndexByte(message, q)
		if start >= 0 {
			end := strings.IndexByte(message[start+1:], q)
			if end > 0 {
				return end
			}
		}
	}

	// Try to extract the last meaningful word (often the problematic token).
	words := strings.Fields(message)
	if len(words) > 0 {
		last := words[len(words)-1]
		// Strip trailing punctuation
		last = strings.TrimRight(last, ".,;:!?")
		if len(last) > 0 {
			return len(last)
		}
	}

	return 1 // Minimum highlight of 1 character
}
