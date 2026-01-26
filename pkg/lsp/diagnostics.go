package lsp

import "github.com/grindlemire/go-tui/pkg/lsp/log"

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	// DiagnosticSeverityError reports an error.
	DiagnosticSeverityError DiagnosticSeverity = 1
	// DiagnosticSeverityWarning reports a warning.
	DiagnosticSeverityWarning DiagnosticSeverity = 2
	// DiagnosticSeverityInformation reports an information.
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	// DiagnosticSeverityHint reports a hint.
	DiagnosticSeverityHint DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic, such as a compiler error or warning.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// PublishDiagnosticsParams represents the parameters for publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// publishDiagnostics sends diagnostics for a document.
func (s *Server) publishDiagnostics(doc *Document) {
	if doc == nil {
		return
	}

	diagnostics := make([]Diagnostic, 0, len(doc.Errors))

	for _, err := range doc.Errors {
		diag := Diagnostic{
			Range:    TuigenPosToRange(err.Pos, estimateErrorLength(err.Message)),
			Severity: DiagnosticSeverityError,
			Source:   "tui",
			Message:  err.Message,
		}
		if err.Hint != "" {
			diag.Message = err.Message + " (" + err.Hint + ")"
		}
		diagnostics = append(diagnostics, diag)
	}

	params := PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diagnostics,
	}

	if err := s.sendNotification("textDocument/publishDiagnostics", params); err != nil {
		log.Server("Error publishing diagnostics: %v", err)
	}
}

// estimateErrorLength estimates the length of text to highlight for an error.
// This is a heuristic since we don't always know the exact span.
func estimateErrorLength(message string) int {
	// Default to highlighting a reasonable chunk
	// In the future, we could parse the error message to find tokens
	return 10
}
