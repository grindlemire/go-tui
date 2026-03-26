package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	logFile *os.File
	mu      sync.Mutex

	overflowOnce      sync.Once
	overflowHighlight bool

	topicsOnce sync.Once
	allTopics  bool            // DEBUG=1 or DEBUG=*
	topics     map[string]bool // DEBUG=keys,dispatch
)

// OverflowHighlight returns true if the TUI_DEBUG_OVERFLOW environment variable
// is set, indicating that containers with overflowing children should be
// highlighted with a bright red background.
func OverflowHighlight() bool {
	overflowOnce.Do(func() {
		overflowHighlight = os.Getenv("TUI_DEBUG_OVERFLOW") != ""
	})
	return overflowHighlight
}

// parseTopics reads the DEBUG env var and populates the topic filter.
// DEBUG=1 or DEBUG=* enables all topics. DEBUG=keys,dispatch enables
// only those topics. Unset or empty disables all logging.
func parseTopics() {
	topicsOnce.Do(func() {
		val := strings.TrimSpace(os.Getenv("DEBUG"))
		if val == "" {
			return
		}
		if val == "1" || val == "*" {
			allTopics = true
			return
		}
		topics = make(map[string]bool)
		for _, t := range strings.Split(val, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				topics[t] = true
			}
		}
	})
}

// topicEnabled returns true if the given topic should be logged.
func topicEnabled(topic string) bool {
	parseTopics()
	if allTopics {
		return true
	}
	return topics[topic]
}

// debugEnabled returns true if any debug logging is configured.
func debugEnabled() bool {
	parseTopics()
	return allTopics || len(topics) > 0
}

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
// Enabled when DEBUG is set to any value.
func Log(format string, args ...any) {
	if !debugEnabled() {
		return
	}

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

// Topic writes a message to the debug log only if the given topic is enabled.
// Topics are enabled via the DEBUG env var: DEBUG=keys,dispatch enables those
// two topics. DEBUG=1 or DEBUG=* enables all topics.
func Topic(topic string, format string, args ...any) {
	if !topicEnabled(topic) {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if logFile == nil {
		initLocked("")
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] [%s] %s\n", timestamp, topic, msg)
	logFile.Sync()
}
