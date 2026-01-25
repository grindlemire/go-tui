package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// runCheck implements the check subcommand.
// It parses and analyzes .tui files without generating code.
// Useful for syntax checking and IDE integration.
func runCheck(args []string) error {
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

	// Collect all .tui files
	files, err := collectTuiFiles(paths)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no .tui files found")
	}

	if verbose {
		fmt.Printf("Checking %d .tui file(s)\n", len(files))
	}

	// Check each file
	var errorCount int
	for _, inputPath := range files {
		if verbose {
			fmt.Printf("Checking %s\n", inputPath)
		}

		if err := checkFile(inputPath); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			errorCount++
			continue
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	if verbose {
		fmt.Printf("All %d file(s) passed checks\n", len(files))
	}

	return nil
}

// checkFile parses and analyzes a single .tui file.
func checkFile(inputPath string) error {
	// Read source file
	source, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Get just the filename for error messages
	filename := filepath.Base(inputPath)

	// Parse source
	lexer := tuigen.NewLexer(filename, string(source))
	parser := tuigen.NewParser(lexer)

	file, parseErr := parser.ParseFile()
	if parseErr != nil {
		return parseErr
	}

	// Analyze (validates elements and attributes)
	analyzer := tuigen.NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		return err
	}

	return nil
}
