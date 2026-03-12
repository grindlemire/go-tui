// Package main demonstrates a foldable directory tree with keyboard navigation.
//
// Usage:
//
//	go run ../../cmd/tui generate tree.gsx
//	go run . [path]
//
// If no path is given, the current directory is used.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate tree.gsx

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	app, err := tui.NewApp(
		tui.WithRootComponent(DirectoryTree(absRoot)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
