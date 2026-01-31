package gopls

import (
	"testing"
)

func TestSourceMapTuiToGo(t *testing.T) {
	type tc struct {
		mappings []Mapping
		tuiLine  int
		tuiCol   int
		wantLine int
		wantCol  int
		wantOk   bool
	}

	tests := map[string]tc{
		"exact match start": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			tuiLine:  5,
			tuiCol:   10,
			wantLine: 10,
			wantCol:  5,
			wantOk:   true,
		},
		"within mapping": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			tuiLine:  5,
			tuiCol:   15, // 5 chars into the mapping
			wantLine: 10,
			wantCol:  10, // 5 + 5 offset
			wantOk:   true,
		},
		"no match": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			tuiLine:  7, // different line
			tuiCol:   10,
			wantLine: 7, // returns original
			wantCol:  10,
			wantOk:   false,
		},
		"before mapping column": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			tuiLine:  5,
			tuiCol:   5, // before mapping starts
			wantLine: 5,
			wantCol:  5,
			wantOk:   false,
		},
		"after mapping column": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			tuiLine:  5,
			tuiCol:   35, // after mapping ends (10 + 20 = 30)
			wantLine: 5,
			wantCol:  35,
			wantOk:   false,
		},
		"empty mappings": {
			mappings: []Mapping{},
			tuiLine:  5,
			tuiCol:   10,
			wantLine: 5,
			wantCol:  10,
			wantOk:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			gotLine, gotCol, gotOk := sm.TuiToGo(tt.tuiLine, tt.tuiCol)

			if gotLine != tt.wantLine {
				t.Errorf("TuiToGo() gotLine = %d, want %d", gotLine, tt.wantLine)
			}
			if gotCol != tt.wantCol {
				t.Errorf("TuiToGo() gotCol = %d, want %d", gotCol, tt.wantCol)
			}
			if gotOk != tt.wantOk {
				t.Errorf("TuiToGo() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestSourceMapGoToTui(t *testing.T) {
	type tc struct {
		mappings []Mapping
		goLine   int
		goCol    int
		wantLine int
		wantCol  int
		wantOk   bool
	}

	tests := map[string]tc{
		"exact match start": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			goLine:   10,
			goCol:    5,
			wantLine: 5,
			wantCol:  10,
			wantOk:   true,
		},
		"within mapping": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			goLine:   10,
			goCol:    15, // 10 chars into the mapping
			wantLine: 5,
			wantCol:  20, // 10 + 10 offset
			wantOk:   true,
		},
		"no match": {
			mappings: []Mapping{
				{TuiLine: 5, TuiCol: 10, GoLine: 10, GoCol: 5, Length: 20},
			},
			goLine:   7, // different line
			goCol:    5,
			wantLine: 7, // returns original
			wantCol:  5,
			wantOk:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			gotLine, gotCol, gotOk := sm.GoToTui(tt.goLine, tt.goCol)

			if gotLine != tt.wantLine {
				t.Errorf("GoToTui() gotLine = %d, want %d", gotLine, tt.wantLine)
			}
			if gotCol != tt.wantCol {
				t.Errorf("GoToTui() gotCol = %d, want %d", gotCol, tt.wantCol)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GoToTui() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestSourceMapRoundTrip(t *testing.T) {
	type tc struct {
		mapping Mapping
		offset  int // offset within the mapping to test
	}

	tests := map[string]tc{
		"start of mapping": {
			mapping: Mapping{TuiLine: 3, TuiCol: 8, GoLine: 7, GoCol: 12, Length: 15},
			offset:  0,
		},
		"middle of mapping": {
			mapping: Mapping{TuiLine: 3, TuiCol: 8, GoLine: 7, GoCol: 12, Length: 15},
			offset:  7,
		},
		"end of mapping": {
			mapping: Mapping{TuiLine: 3, TuiCol: 8, GoLine: 7, GoCol: 12, Length: 15},
			offset:  14, // Length - 1
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			sm.AddMapping(tt.mapping)

			// Start from TUI position
			tuiLine := tt.mapping.TuiLine
			tuiCol := tt.mapping.TuiCol + tt.offset

			// Convert to Go
			goLine, goCol, ok := sm.TuiToGo(tuiLine, tuiCol)
			if !ok {
				t.Fatalf("TuiToGo failed to find mapping")
			}

			// Convert back to TUI
			backLine, backCol, ok := sm.GoToTui(goLine, goCol)
			if !ok {
				t.Fatalf("GoToTui failed to find mapping")
			}

			// Should match original
			if backLine != tuiLine || backCol != tuiCol {
				t.Errorf("Round trip failed: started at (%d, %d), got back (%d, %d)",
					tuiLine, tuiCol, backLine, backCol)
			}
		})
	}
}

func TestVirtualFileCache(t *testing.T) {
	type tc struct {
		operations func(c *VirtualFileCache)
		tuiURI     string
		wantFound  bool
	}

	tests := map[string]tc{
		"put and get": {
			operations: func(c *VirtualFileCache) {
				c.Put("file:///test.gsx", "file:///test_gsx_generated.go", "content", NewSourceMap(), 1)
			},
			tuiURI:    "file:///test.gsx",
			wantFound: true,
		},
		"get nonexistent": {
			operations: func(c *VirtualFileCache) {},
			tuiURI:     "file:///nonexistent.gsx",
			wantFound:  false,
		},
		"put and remove": {
			operations: func(c *VirtualFileCache) {
				c.Put("file:///test.gsx", "file:///test_gsx_generated.go", "content", NewSourceMap(), 1)
				c.Remove("file:///test.gsx")
			},
			tuiURI:    "file:///test.gsx",
			wantFound: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cache := NewVirtualFileCache()
			tt.operations(cache)

			got := cache.Get(tt.tuiURI)
			if tt.wantFound && got == nil {
				t.Error("expected to find cached file, got nil")
			}
			if !tt.wantFound && got != nil {
				t.Errorf("expected not to find cached file, got %+v", got)
			}
		})
	}
}

func TestVirtualFileCacheGetByGoURI(t *testing.T) {
	cache := NewVirtualFileCache()
	sm := NewSourceMap()

	cache.Put("file:///a.gsx", "file:///a_gsx_generated.go", "content a", sm, 1)
	cache.Put("file:///b.gsx", "file:///b_gsx_generated.go", "content b", sm, 1)

	// Find by Go URI
	got := cache.GetByGoURI("file:///a_gsx_generated.go")
	if got == nil {
		t.Fatal("expected to find cached file by Go URI")
	}
	if got.TuiURI != "file:///a.gsx" {
		t.Errorf("got TuiURI %s, want file:///a.gsx", got.TuiURI)
	}

	// Find nonexistent
	got = cache.GetByGoURI("file:///nonexistent_gsx_generated.go")
	if got != nil {
		t.Error("expected nil for nonexistent Go URI")
	}
}

func TestTuiURIToGoURI(t *testing.T) {
	type tc struct {
		tuiURI  string
		wantURI string
	}

	tests := map[string]tc{
		"gsx extension": {
			tuiURI:  "file:///path/to/file.gsx",
			wantURI: "file:///path/to/file_gsx_generated.go",
		},
		"no extension": {
			tuiURI:  "file:///path/to/file",
			wantURI: "file:///path/to/file_generated.go",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := TuiURIToGoURI(tt.tuiURI)
			if got != tt.wantURI {
				t.Errorf("TuiURIToGoURI(%q) = %q, want %q", tt.tuiURI, got, tt.wantURI)
			}
		})
	}
}

func TestGoURIToTuiURI(t *testing.T) {
	type tc struct {
		goURI   string
		wantURI string
	}

	tests := map[string]tc{
		"generated suffix": {
			goURI:   "file:///path/to/file_gsx_generated.go",
			wantURI: "file:///path/to/file.gsx",
		},
		"no suffix": {
			goURI:   "file:///path/to/regular.go",
			wantURI: "file:///path/to/regular.go",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := GoURIToTuiURI(tt.goURI)
			if got != tt.wantURI {
				t.Errorf("GoURIToTuiURI(%q) = %q, want %q", tt.goURI, got, tt.wantURI)
			}
		})
	}
}

func TestIsVirtualGoFile(t *testing.T) {
	type tc struct {
		uri  string
		want bool
	}

	tests := map[string]tc{
		"virtual file":  {uri: "file:///test_gsx_generated.go", want: true},
		"regular go":    {uri: "file:///test.go", want: false},
		"gsx file":      {uri: "file:///test.gsx", want: false},
		"almost suffix": {uri: "file:///test_generated.go", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsVirtualGoFile(tt.uri)
			if got != tt.want {
				t.Errorf("IsVirtualGoFile(%q) = %v, want %v", tt.uri, got, tt.want)
			}
		})
	}
}

