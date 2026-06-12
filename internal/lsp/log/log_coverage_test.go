package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withLogFile points the package logger at a fresh temp file, runs fn, then
// restores the disabled state and returns the file contents.
func withLogFile(t *testing.T, fn func()) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "lsp.log")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating log file: %v", err)
	}

	SetOutput(f)
	defer SetOutput(nil)

	fn()

	if err := f.Close(); err != nil {
		t.Fatalf("closing log file: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	return string(data)
}

func TestLogFunctionsWriteWithPrefixes(t *testing.T) {
	type tc struct {
		log  func(format string, args ...any)
		want string
	}

	tests := map[string]tc{
		"Debug has no prefix": {
			log:  Debug,
			want: "value=42\n",
		},
		"Debugf aliases Debug": {
			log:  Debugf,
			want: "value=42\n",
		},
		"Server prefix": {
			log:  Server,
			want: "[server] value=42\n",
		},
		"Gopls prefix": {
			log:  Gopls,
			want: "[gopls] value=42\n",
		},
		"Generate prefix": {
			log:  Generate,
			want: "[generate] value=42\n",
		},
		"Mapping prefix": {
			log:  Mapping,
			want: "[mapping] value=42\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := withLogFile(t, func() {
				tt.log("value=%d", 42)
			})
			if got != tt.want {
				t.Errorf("log output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogDisabledWritesNothing(t *testing.T) {
	// Logging is disabled by default (no SetOutput call). All log functions
	// must be no-ops, which we verify by calling them and then enabling a
	// real file to prove the earlier calls did not land anywhere.
	SetOutput(nil)

	if Enabled() {
		t.Fatal("Enabled() = true before SetOutput, want false")
	}

	// These must not panic and must not write anywhere.
	Debug("dropped %s", "debug")
	Debugf("dropped %s", "debugf")
	Server("dropped %s", "server")
	Gopls("dropped %s", "gopls")
	Generate("dropped %s", "generate")
	Mapping("dropped %s", "mapping")

	got := withLogFile(t, func() {
		Debug("kept")
	})
	if strings.Contains(got, "dropped") {
		t.Errorf("disabled log calls leaked into file: %q", got)
	}
	if got != "kept\n" {
		t.Errorf("log output = %q, want %q", got, "kept\n")
	}
}

func TestEnabledReflectsSetOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "enabled.log")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating log file: %v", err)
	}
	defer f.Close()

	SetOutput(f)
	if !Enabled() {
		t.Error("Enabled() = false after SetOutput(file), want true")
	}

	SetOutput(nil)
	if Enabled() {
		t.Error("Enabled() = true after SetOutput(nil), want false")
	}
}
