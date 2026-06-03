package lsp

import (
	"testing"
)

func TestComponentIndex(t *testing.T) {
	type tc struct {
		content      string
		wantComps    []string
		lookupName   string
		lookupExists bool
	}

	tests := map[string]tc{
		"single component": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantComps:    []string{"Hello"},
			lookupName:   "Hello",
			lookupExists: true,
		},
		"multiple components": {
			content: `package main

templ Header() {
	<span>Header</span>
}

templ Footer() {
	<span>Footer</span>
}
`,
			wantComps:    []string{"Header", "Footer"},
			lookupName:   "Footer",
			lookupExists: true,
		},
		"lookup nonexistent": {
			content: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantComps:    []string{"Hello"},
			lookupName:   "NotExists",
			lookupExists: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dm := NewDocumentManager()
			idx := NewComponentIndex()

			uri := "file:///test.gsx"
			doc := dm.Open(uri, tt.content, 1)

			idx.IndexDocument(uri, doc.AST)

			// Check all expected components are indexed
			for _, compName := range tt.wantComps {
				if _, ok := idx.Lookup(compName); !ok {
					t.Errorf("expected component %s to be indexed", compName)
				}
			}

			// Test lookup
			_, exists := idx.Lookup(tt.lookupName)
			if exists != tt.lookupExists {
				t.Errorf("Lookup(%s) = _, %v; want _, %v", tt.lookupName, exists, tt.lookupExists)
			}
		})
	}
}

func TestComponentIndexRemove(t *testing.T) {
	dm := NewDocumentManager()
	idx := NewComponentIndex()

	uri := "file:///test.gsx"
	content := `package main

templ Hello() {
	<span>Hello</span>
}
`
	doc := dm.Open(uri, content, 1)
	idx.IndexDocument(uri, doc.AST)

	// Verify component is indexed
	if _, ok := idx.Lookup("Hello"); !ok {
		t.Fatal("expected Hello to be indexed")
	}

	// Remove the file
	idx.Remove(uri)

	// Verify component is removed
	if _, ok := idx.Lookup("Hello"); ok {
		t.Fatal("expected Hello to be removed from index")
	}
}
