package lsp

import (
	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/lsp/provider"
)

// Diagnostic and DiagnosticSeverity are type aliases for the canonical definitions
// in the provider package, eliminating duplicate type definitions.
type Diagnostic = provider.Diagnostic
type DiagnosticSeverity = provider.DiagnosticSeverity

// Re-export severity constants so existing lsp package code compiles unchanged.
const (
	DiagnosticSeverityError       = provider.DiagnosticSeverityError
	DiagnosticSeverityWarning     = provider.DiagnosticSeverityWarning
	DiagnosticSeverityInformation = provider.DiagnosticSeverityInformation
	DiagnosticSeverityHint        = provider.DiagnosticSeverityHint
)

// PublishDiagnosticsParams represents the parameters for publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// publishDiagnostics sends diagnostics for a document.
// If a DiagnosticsProvider is registered, it delegates to the provider;
// otherwise it falls back to inline conversion.
// Gopls diagnostics are merged with parse diagnostics.
func (s *Server) publishDiagnostics(doc *Document) {
	if doc == nil {
		return
	}

	var diagnostics []Diagnostic

	if s.router != nil && s.router.registry != nil && s.router.registry.Diagnostics != nil {
		diags, err := s.router.registry.Diagnostics.Diagnose(doc)
		if err != nil {
			log.Server("Diagnostics provider error: %v", err)
			diagnostics = []Diagnostic{}
		} else {
			diagnostics = diags
		}
	} else {
		// No provider registered — fall back to inline conversion of parse errors
		for _, e := range doc.Errors {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: e.Pos.Line - 1, Character: e.Pos.Column - 1},
					End:   Position{Line: e.Pos.Line - 1, Character: e.Pos.Column - 1 + 10},
				},
				Severity: DiagnosticSeverityError,
				Source:   "gsx",
				Message:  e.Message,
			})
		}
	}

	// Add gopls diagnostics (type errors, undefined identifiers, etc.)
	s.goplsDiagnosticsMu.RLock()
	goplsDiags := s.goplsDiagnostics[doc.URI]
	s.goplsDiagnosticsMu.RUnlock()

	for _, gd := range goplsDiags {
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: gd.Range.Start.Line, Character: gd.Range.Start.Character},
				End:   Position{Line: gd.Range.End.Line, Character: gd.Range.End.Character},
			},
			Severity: DiagnosticSeverity(gd.Severity),
			Source:   gd.Source,
			Message:  gd.Message,
		})
	}

	params := PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diagnostics,
	}

	if err := s.sendNotification("textDocument/publishDiagnostics", params); err != nil {
		log.Server("Error publishing diagnostics: %v", err)
	}
}
