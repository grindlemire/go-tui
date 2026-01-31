package lsp

import (
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/lsp/provider"
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
		diagnostics = []Diagnostic{}
	}

	params := PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diagnostics,
	}

	if err := s.sendNotification("textDocument/publishDiagnostics", params); err != nil {
		log.Server("Error publishing diagnostics: %v", err)
	}
}
