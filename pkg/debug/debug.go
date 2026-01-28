package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex
)

// Init initializes debug logging to the specified file path.
// If path is empty, uses "debug.log" in the current directory.
func Init(path string) error {
	mu.Lock()
	defer mu.Unlock()
	return initLocked(path)
}

// initLocked does the actual init work. Caller must hold mu.
func initLocked(path string) error {
	if path == "" {
		path = "debug.log"
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open debug log: %w", err)
	}

	logFile = f
	return nil
}

// Close closes the debug log file.
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}

// Log writes a message to the debug log with a timestamp.
func Log(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()

	if logFile == nil {
		initLocked("")
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
	logFile.Sync()
}

// Logf is an alias for Log.
func Logf(format string, args ...any) {
	Log(format, args...)
}
