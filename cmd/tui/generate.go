package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// runGenerate implements the generate subcommand.
// It processes .gsx files and generates corresponding Go source files.
func runGenerate(args []string) error {
	verbose := false
	var paths []string

	// Parse arguments
	for _, arg := range args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else {
			paths = append(paths, arg)
		}
	}

	// Default to current directory if no paths specified
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Collect all .gsx files
	files, err := collectGsxFiles(paths)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no .gsx files found")
	}

	if verbose {
		fmt.Printf("Found %d .gsx file(s)\n", len(files))
	}

	// Process each file
	var errorCount int
	for _, inputPath := range files {
		outputPath := outputFileName(inputPath)

		if verbose {
			fmt.Printf("Processing %s -> %s\n", inputPath, outputPath)
		}

		if err := generateFile(inputPath, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", inputPath, err)
			errorCount++
			continue
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	if verbose {
		fmt.Printf("Successfully generated %d file(s)\n", len(files))
	}

	return nil
}

// collectGsxFiles finds all .gsx files from the given paths.
// Supports:
//   - Direct file paths: "header.gsx"
//   - Directory paths: "./components"
//   - Recursive pattern: "./..."
func collectGsxFiles(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		// Handle ./... recursive pattern
		if strings.HasSuffix(path, "/...") {
			root := strings.TrimSuffix(path, "/...")
			if root == "." || root == "" {
				root = "."
			}

			err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(p, ".gsx") {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking %s: %w", root, err)
			}
			continue
		}

		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}

		if info.IsDir() {
			// Collect all .gsx files in directory (non-recursive)
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("reading directory %s: %w", path, err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gsx") {
					files = append(files, filepath.Join(path, entry.Name()))
				}
			}
		} else if strings.HasSuffix(path, ".gsx") {
			files = append(files, path)
		}
	}

	return files, nil
}

// outputFileName converts a .gsx filename to its output .go filename.
// Examples:
//
//	header.gsx     -> header_gsx.go
//	my-app.gsx     -> my_app_gsx.go
//	components.gsx -> components_gsx.go
func outputFileName(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)

	// Remove .gsx extension
	name := strings.TrimSuffix(base, ".gsx")

	// Replace hyphens with underscores (Go doesn't like hyphens in filenames)
	name = strings.ReplaceAll(name, "-", "_")

	// Add _gsx.go suffix
	output := name + "_gsx.go"

	return filepath.Join(dir, output)
}

// generateFile parses a .gsx file and generates the corresponding Go file.
func generateFile(inputPath, outputPath string) error {
	// Read source file
	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Get just the filename for error messages and header comment
	filename := filepath.Base(inputPath)

	// Parse source
	lexer := tuigen.NewLexer(filename, string(source))
	parser := tuigen.NewParser(lexer)

	file, err := parser.ParseFile()
	if err != nil {
		return err
	}

	// Analyze (validates and adds missing imports)
	analyzer := tuigen.NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		return err
	}

	// Generate Go code
	generator := tuigen.NewGenerator()
	output, err := generator.Generate(file, filename)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
