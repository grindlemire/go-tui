package tree_sitter_tui_test

import (
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/tree-sitter/tree-sitter-tui"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_tui.Language())
	if language == nil {
		t.Errorf("Error loading Tui grammar")
	}
}
