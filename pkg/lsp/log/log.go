// Package log provides centralized logging for the LSP server.
package log

import (
	"fmt"
	"os"
	"sync"
)

var (
	file *os.File
	mu   sync.Mutex
)

// SetOutput sets the log output file. Pass nil to disable logging.
func SetOutput(f *os.File) {
	mu.Lock()
	defer mu.Unlock()
	file = f
}

// Debug writes a debug log message if logging is enabled.
func Debug(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		fmt.Fprintf(file, format+"\n", args...)
	}
}

// Debugf is an alias for Debug.
func Debugf(format string, args ...any) {
	Debug(format, args...)
}

// Server writes a server-prefixed log message.
func Server(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		fmt.Fprintf(file, "[server] "+format+"\n", args...)
	}
}

// Gopls writes a gopls-prefixed log message.
func Gopls(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		fmt.Fprintf(file, "[gopls] "+format+"\n", args...)
	}
}

// Generate writes a generate-prefixed log message.
func Generate(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		fmt.Fprintf(file, "[generate] "+format+"\n", args...)
	}
}

// Mapping writes a mapping-prefixed log message.
func Mapping(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		fmt.Fprintf(file, "[mapping] "+format+"\n", args...)
	}
}

// Enabled returns true if logging is enabled.
func Enabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return file != nil
}
