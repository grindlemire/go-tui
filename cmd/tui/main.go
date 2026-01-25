// Package main provides the CLI tool for the .tui DSL compiler.
//
// Usage:
//
//	tui generate [path...]    Generate Go code from .tui files
//	tui check [path...]       Check .tui files without generating
//	tui help                  Show help
//
// Examples:
//
//	tui generate ./...        Recursively find and compile all .tui files
//	tui generate ./components Process a specific directory
//	tui generate header.tui   Process a specific file
//	tui check header.tui      Check syntax without generating
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

const usage = `tui - DSL compiler for go-tui element trees

Usage:
  tui <command> [options] [path...]

Commands:
  generate    Generate Go code from .tui files
  check       Check .tui files without generating code
  fmt         Format .tui files
  lsp         Start the language server (for editor integration)
  version     Print version information
  help        Show this help message

Options:
  -v          Verbose output

Examples:
  tui generate ./...              Recursively process all .tui files
  tui generate ./components       Process files in a directory
  tui generate header.tui         Process a specific file
  tui generate -v ./...           Verbose output during generation
  tui check header.tui            Check syntax without generating
  tui fmt ./...                   Format all .tui files recursively
  tui fmt --check ./...           Check formatting without modifying
  tui fmt --stdout file.tui       Print formatted output to stdout
  tui lsp                         Start LSP server on stdio
  tui lsp --log /tmp/tui-lsp.log  Start with debug logging

For more information, see https://github.com/grindlemire/go-tui
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "generate":
		if err := runGenerate(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := runCheck(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "fmt":
		if err := runFmt(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "lsp":
		if err := runLSP(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("tui version %s\n", version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}
