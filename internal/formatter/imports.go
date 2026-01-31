package formatter

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// fixImports generates Go code from the AST, runs goimports to resolve
// missing imports, then updates the AST with the corrected imports.
func fixImports(file *tuigen.File, filename string) error {
	// Convert .tui filename to .go for goimports resolution
	goFilename := filename
	if strings.HasSuffix(goFilename, ".tui") {
		goFilename = strings.TrimSuffix(goFilename, ".tui") + "_tui.go"
	}

	// Generate Go code from the AST (generator runs imports.Process internally)
	gen := tuigen.NewGenerator()
	goCode, err := gen.Generate(file, goFilename)
	if err != nil {
		// If generation fails, skip import fixing but don't fail formatting
		return nil
	}

	// The generator already runs imports.Process, so goCode has correct imports.
	// Parse the imports from the generated Go code.
	newImports, err := extractImports(goCode)
	if err != nil {
		return nil
	}

	// Update the AST with the new imports
	file.Imports = newImports
	return nil
}

// extractImports parses Go source code and extracts import declarations.
func extractImports(goCode []byte) ([]tuigen.Import, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", goCode, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var result []tuigen.Import
	for _, imp := range f.Imports {
		ti := tuigen.Import{
			Path: imp.Path.Value[1 : len(imp.Path.Value)-1], // Remove quotes
		}
		if imp.Name != nil {
			ti.Alias = imp.Name.Name
		}
		result = append(result, ti)
	}

	return result, nil
}
