package tree_sitter_gsx_test

import (
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/tree-sitter/tree-sitter-gsx"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_gsx.Language())
	if language == nil {
		t.Errorf("Error loading GSX grammar")
	}
}
