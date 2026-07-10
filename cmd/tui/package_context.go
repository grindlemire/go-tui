package main

import (
	"path/filepath"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// loadPackageContext builds a PackageContext from the sibling files of
// inputPath's directory, so generation and analysis can detect user code
// declared outside the .gsx file itself.
func loadPackageContext(inputPath string) *tuigen.PackageContext {
	ctx := tuigen.NewPackageContext()
	self := filepath.Base(inputPath)
	ctx.AddDirectory(filepath.Dir(inputPath), func(filename string) bool {
		return filename == self
	})
	return ctx
}
