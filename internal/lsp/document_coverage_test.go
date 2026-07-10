package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestDocumentManagerUpdate_UnopenedDocument(t *testing.T) {
	dm := NewDocumentManager()
	uri := "file:///fresh.gsx"

	doc := dm.Update(uri, "package main\n\ntempl Fresh() {\n\t<span>hi</span>\n}\n", 3)
	if doc == nil {
		t.Fatal("Update returned nil")
	}
	if doc.Version != 3 {
		t.Errorf("Version = %d, want 3", doc.Version)
	}
	if doc.AST == nil {
		t.Error("expected document to be parsed")
	}
	if dm.Get(uri) != doc {
		t.Error("document not registered in the manager")
	}
}

// TestDocumentManager_CrossFileCollisions verifies that analysis of one
// document sees declarations from sibling files: open documents through
// their in-memory ASTs (unsaved edits included), unopened files from disk.
func TestDocumentManager_CrossFileCollisions(t *testing.T) {
	dm := NewDocumentManager()

	// Open a sibling that declares templ Foo. There is no disk file at all,
	// proving detection went through the open-buffer path.
	dm.Open("file:///ws/a.gsx", "package x\n\ntempl Foo() {\n\t<span>a</span>\n}\n", 1)

	doc := dm.Open("file:///ws/b.gsx", "package x\n\ntempl Foo() {\n\t<span>b</span>\n}\n", 1)

	found := false
	for _, e := range doc.Errors {
		if strings.Contains(e.Message, "duplicate templ") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate templ error from open sibling, got %v", doc.Errors)
	}

	// A document in a different directory must not collide.
	other := dm.Open("file:///elsewhere/c.gsx", "package x\n\ntempl Foo() {\n\t<span>c</span>\n}\n", 1)
	for _, e := range other.Errors {
		if strings.Contains(e.Message, "duplicate templ") {
			t.Errorf("unexpected collision across directories: %v", e.Message)
		}
	}

	// Unopened sibling .go file on disk is picked up.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "helpers.go"), []byte("package x\n\nfunc Bar() string { return \"x\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	diskDoc := dm.Open("file://"+filepath.Join(dir, "d.gsx"), "package x\n\ntempl Bar() {\n\t<span>d</span>\n}\n", 1)
	found = false
	for _, e := range diskDoc.Errors {
		if strings.Contains(e.Message, "conflicts with a Go function") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected collision with on-disk sibling function, got %v", diskDoc.Errors)
	}
}

func TestURIToPath(t *testing.T) {
	type tc struct {
		uri  string
		want string
	}

	tests := map[string]tc{
		"file uri":        {uri: "file:///tmp/a.gsx", want: "/tmp/a.gsx"},
		"plain path":      {uri: "relative/a.gsx", want: "relative/a.gsx"},
		"bare prefix":     {uri: "file://", want: "file://"},
		"non-file scheme": {uri: "untitled:Untitled-1", want: "untitled:Untitled-1"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := uriToPath(tt.uri); got != tt.want {
				t.Errorf("uriToPath(%q) = %q, want %q", tt.uri, got, tt.want)
			}
		})
	}
}

func TestPositionToOffset_PastLineEnd(t *testing.T) {
	// Requesting a character beyond the line end falls back to start + character.
	content := "ab\ncd"
	got := PositionToOffset(content, Position{Line: 0, Character: 10})
	if got != 10 {
		t.Errorf("PositionToOffset = %d, want 10", got)
	}
}

func TestTuigenPosToRange(t *testing.T) {
	got := TuigenPosToRange(tuigen.Position{Line: 3, Column: 5}, 4)
	want := Range{
		Start: Position{Line: 2, Character: 4},
		End:   Position{Line: 2, Character: 8},
	}
	if got != want {
		t.Errorf("TuigenPosToRange = %+v, want %+v", got, want)
	}
}

func TestTuigenPosToRangeWithEnd(t *testing.T) {
	got := TuigenPosToRangeWithEnd(
		tuigen.Position{Line: 1, Column: 2},
		tuigen.Position{Line: 4, Column: 7},
	)
	want := Range{
		Start: Position{Line: 0, Character: 1},
		End:   Position{Line: 3, Character: 6},
	}
	if got != want {
		t.Errorf("TuigenPosToRangeWithEnd = %+v, want %+v", got, want)
	}
}
