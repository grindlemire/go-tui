package gopls

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

func TestGenerateVirtualGo(t *testing.T) {
	type tc struct {
		file           *tuigen.File
		wantContains   []string
		wantMinMapLen  int
	}

	tests := map[string]tc{
		"simple component": {
			file: &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name:   "Hello",
						Params: []*tuigen.Param{},
						Body: []tuigen.Node{
							&tuigen.Element{
								Tag: "text",
								Children: []tuigen.Node{
									&tuigen.GoExpr{
										Code: `"hello"`,
										Position: tuigen.Position{
											Line:   4,
											Column: 8,
										},
									},
								},
							},
						},
					},
				},
			},
			wantContains: []string{
				"package main",
				"func Hello()",
				`_ = "hello"`,
				"return nil",
			},
			wantMinMapLen: 1, // at least one mapping for the Go expression
		},
		"component with params": {
			file: &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name: "Counter",
						Params: []*tuigen.Param{
							{Name: "count", Type: "int"},
							{Name: "label", Type: "string"},
						},
						Body: []tuigen.Node{},
					},
				},
			},
			wantContains: []string{
				"package main",
				"func Counter(count int, label string)",
			},
			wantMinMapLen: 0,
		},
		"component with for loop": {
			file: &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name: "List",
						Body: []tuigen.Node{
							&tuigen.ForLoop{
								Index:    "i",
								Value:    "item",
								Iterable: "items",
								Position: tuigen.Position{Line: 3, Column: 2},
								Body: []tuigen.Node{
									&tuigen.Element{Tag: "text"},
								},
							},
						},
					},
				},
			},
			wantContains: []string{
				"for i, item := range items",
			},
			wantMinMapLen: 1, // mapping for the iterable
		},
		"component with if statement": {
			file: &tuigen.File{
				Package: "main",
				Components: []*tuigen.Component{
					{
						Name: "Toggle",
						Body: []tuigen.Node{
							&tuigen.IfStmt{
								Condition: "show",
								Position:  tuigen.Position{Line: 3, Column: 2},
								Then: []tuigen.Node{
									&tuigen.Element{Tag: "text"},
								},
							},
						},
					},
				},
			},
			wantContains: []string{
				"if show {",
			},
			wantMinMapLen: 1, // mapping for the condition
		},
		"with imports": {
			file: &tuigen.File{
				Package: "main",
				Imports: []tuigen.Import{
					{Path: "fmt"},
					{Alias: "el", Path: "github.com/grindlemire/go-tui/internal/element"},
				},
				Components: []*tuigen.Component{
					{Name: "Test", Body: []tuigen.Node{}},
				},
			},
			wantContains: []string{
				"package main",
				`"fmt"`,
				`el "github.com/grindlemire/go-tui/internal/element"`,
			},
			wantMinMapLen: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			goContent, sourceMap := GenerateVirtualGo(tt.file)

			for _, want := range tt.wantContains {
				if !containsString(goContent, want) {
					t.Errorf("generated Go code does not contain %q:\n%s", want, goContent)
				}
			}

			if sourceMap.Len() < tt.wantMinMapLen {
				t.Errorf("expected at least %d mappings, got %d", tt.wantMinMapLen, sourceMap.Len())
			}
		})
	}
}

func TestGenerateVirtualGoWithComponentCall(t *testing.T) {
	file := &tuigen.File{
		Package: "main",
		Components: []*tuigen.Component{
			{
				Name: "App",
				Body: []tuigen.Node{
					&tuigen.ComponentCall{
						Name:     "Header",
						Args:     `"title"`,
						Position: tuigen.Position{Line: 3, Column: 2},
					},
				},
			},
		},
	}

	goContent, sourceMap := GenerateVirtualGo(file)

	// Should contain the component call as a function call
	if !containsString(goContent, `_ = Header("title")`) {
		t.Errorf("expected component call in generated code:\n%s", goContent)
	}

	// Should have a mapping for the args
	if sourceMap.Len() < 1 {
		t.Error("expected at least one mapping for component call args")
	}
}

func TestSourceMapIsInGoExpression(t *testing.T) {
	sm := NewSourceMap()
	sm.AddMapping(Mapping{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20})

	type tc struct {
		line int
		col  int
		want bool
	}

	tests := map[string]tc{
		"inside expression": {
			line: 5,
			col:  15,
			want: true,
		},
		"at expression start": {
			line: 5,
			col:  10,
			want: true,
		},
		"before expression": {
			line: 5,
			col:  5,
			want: false,
		},
		"after expression": {
			line: 5,
			col:  35,
			want: false,
		},
		"different line": {
			line: 6,
			col:  15,
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := sm.IsInGoExpression(tt.line, tt.col)
			if got != tt.want {
				t.Errorf("IsInGoExpression(%d, %d) = %v, want %v", tt.line, tt.col, got, tt.want)
			}
		})
	}
}

func TestSourceMapAllMappings(t *testing.T) {
	sm := NewSourceMap()

	// Add mappings out of order
	sm.AddMapping(Mapping{TuiLine: 10, TuiCol: 5, GoLine: 20, GoCol: 5, Length: 10})
	sm.AddMapping(Mapping{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 10})
	sm.AddMapping(Mapping{TuiLine: 5, TuiCol: 5, GoLine: 10, GoCol: 10, Length: 5})

	all := sm.AllMappings()

	if len(all) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(all))
	}

	// Should be sorted by TUI position
	if all[0].TuiLine != 5 || all[0].TuiCol != 5 {
		t.Errorf("first mapping should be (5,5), got (%d,%d)", all[0].TuiLine, all[0].TuiCol)
	}
	if all[1].TuiLine != 5 || all[1].TuiCol != 10 {
		t.Errorf("second mapping should be (5,10), got (%d,%d)", all[1].TuiLine, all[1].TuiCol)
	}
	if all[2].TuiLine != 10 {
		t.Errorf("third mapping should have line 10, got %d", all[2].TuiLine)
	}
}

func TestSourceMapClear(t *testing.T) {
	sm := NewSourceMap()
	sm.AddMapping(Mapping{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20})

	if sm.Len() != 1 {
		t.Fatalf("expected 1 mapping, got %d", sm.Len())
	}

	sm.Clear()

	if sm.Len() != 0 {
		t.Errorf("expected 0 mappings after clear, got %d", sm.Len())
	}
}

// containsString checks if haystack contains needle.
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && findString(haystack, needle)
}

func findString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
