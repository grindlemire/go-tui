package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/grindlemire/go-tui/pkg/formatter"
)

// runFmt implements the fmt subcommand.
// It formats .gsx files in place or checks formatting.
func runFmt(args []string) error {
	var (
		stdout bool // print to stdout instead of modifying file
		check  bool // check mode (exit 1 if not formatted)
		paths  []string
	)

	// Parse arguments
	for _, arg := range args {
		switch arg {
		case "--stdout", "-stdout":
			stdout = true
		case "--check", "-check":
			check = true
		default:
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

	// Process files
	fmtr := formatter.New()

	if check {
		return runFmtCheck(fmtr, files)
	}

	if stdout {
		return runFmtStdout(fmtr, files)
	}

	return runFmtInPlace(fmtr, files)
}

// runFmtInPlace formats files in place, modifying them on disk.
func runFmtInPlace(fmtr *formatter.Formatter, files []string) error {
	type result struct {
		path    string
		changed bool
		err     error
	}

	results := make(chan result, len(files))
	var wg sync.WaitGroup

	// Process files in parallel
	for _, path := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			source, err := os.ReadFile(p)
			if err != nil {
				results <- result{path: p, err: fmt.Errorf("reading file: %w", err)}
				return
			}

			res, err := fmtr.FormatWithResult(filepath.Base(p), string(source))
			if err != nil {
				results <- result{path: p, err: err}
				return
			}

			if res.Changed {
				if err := os.WriteFile(p, []byte(res.Content), 0644); err != nil {
					results <- result{path: p, err: fmt.Errorf("writing file: %w", err)}
					return
				}
			}

			results <- result{path: p, changed: res.Changed}
		}(path)
	}

	// Wait for all goroutines and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errorCount int
	for res := range results {
		if res.err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", res.path, res.err)
			errorCount++
		} else if res.changed {
			fmt.Printf("Formatted: %s\n", res.path)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	return nil
}

// runFmtStdout formats files and prints to stdout.
func runFmtStdout(fmtr *formatter.Formatter, files []string) error {
	var errorCount int

	for _, path := range files {
		source, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: reading file: %v\n", path, err)
			errorCount++
			continue
		}

		formatted, err := fmtr.Format(filepath.Base(path), string(source))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			errorCount++
			continue
		}

		if len(files) > 1 {
			fmt.Printf("// %s\n", path)
		}
		fmt.Print(formatted)
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	return nil
}

// runFmtCheck checks if files are formatted without modifying them.
// Returns an error if any file is not formatted.
func runFmtCheck(fmtr *formatter.Formatter, files []string) error {
	type result struct {
		path       string
		notFormatted bool
		err        error
	}

	results := make(chan result, len(files))
	var wg sync.WaitGroup

	// Process files in parallel
	for _, path := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			source, err := os.ReadFile(p)
			if err != nil {
				results <- result{path: p, err: fmt.Errorf("reading file: %w", err)}
				return
			}

			res, err := fmtr.FormatWithResult(filepath.Base(p), string(source))
			if err != nil {
				results <- result{path: p, err: err}
				return
			}

			results <- result{path: p, notFormatted: res.Changed}
		}(path)
	}

	// Wait for all goroutines and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errorCount, notFormattedCount int
	for res := range results {
		if res.err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", res.path, res.err)
			errorCount++
		} else if res.notFormatted {
			fmt.Fprintf(os.Stderr, "ERROR: %s is not formatted\n", res.path)
			notFormattedCount++
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) had errors", errorCount)
	}

	if notFormattedCount > 0 {
		return fmt.Errorf("%d file(s) not formatted", notFormattedCount)
	}

	return nil
}
