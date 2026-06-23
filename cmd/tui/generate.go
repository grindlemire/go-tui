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
	outputPath := ""
	var paths []string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--verbose":
			verbose = true
		case "-o", "--output":
			if i+1 >= len(args) {
				return fmt.Errorf("missing value for %s", args[i])
			}
			outputPath = args[i+1]
			i++ // skip the value we just consumed
		default:
			paths = append(paths, args[i])
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
		finalPath := outputFileName(inputPath, outputPath)

		if verbose {
			fmt.Printf("Processing %s -> %s\n", inputPath, finalPath)
		}

		if err := generateFile(inputPath, finalPath); err != nil {
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
		if before, ok := strings.CutSuffix(path, "/..."); ok {
			root := before
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
func outputFileName(inputPath, outputDir string) string {
	// outputDir is an explicit target directory; otherwise mirror the input file.
	dir := outputDir
	if dir == "" {
		dir = filepath.Dir(inputPath)
	}

	// Strip .gsx and replace hyphens (Go doesn't like hyphens in filenames).
	name := strings.TrimSuffix(filepath.Base(inputPath), ".gsx")
	name = strings.ReplaceAll(name, "-", "_")

	// filepath.Join cleans any trailing slash on dir for us.
	return filepath.Join(dir, name+"_gsx.go")
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

	// Ensure the output directory exists (MkdirAll is a no-op if it already does).
	if dir := filepath.Dir(outputPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(outputPath, output, 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
