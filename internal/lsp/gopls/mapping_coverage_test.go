package gopls

import (
	"testing"
)

func TestFindMappingForGoPosition(t *testing.T) {
	type tc struct {
		mappings []Mapping
		goLine   int
		goCol    int
		want     *Mapping
	}

	m1 := Mapping{TuiLine: 2, TuiCol: 8, GoLine: 10, GoCol: 5, Length: 6}
	m2 := Mapping{TuiLine: 4, TuiCol: 1, GoLine: 12, GoCol: 0, Length: 3}

	tests := map[string]tc{
		"exact start": {
			mappings: []Mapping{m1, m2},
			goLine:   10,
			goCol:    5,
			want:     &m1,
		},
		"inclusive exclusive end boundary": {
			mappings: []Mapping{m1, m2},
			goLine:   10,
			goCol:    11, // GoCol + Length
			want:     &m1,
		},
		"second mapping": {
			mappings: []Mapping{m1, m2},
			goLine:   12,
			goCol:    2,
			want:     &m2,
		},
		"wrong line": {
			mappings: []Mapping{m1, m2},
			goLine:   11,
			goCol:    5,
			want:     nil,
		},
		"before mapping start": {
			mappings: []Mapping{m1},
			goLine:   10,
			goCol:    4,
			want:     nil,
		},
		"past exclusive end": {
			mappings: []Mapping{m1},
			goLine:   10,
			goCol:    12,
			want:     nil,
		},
		"empty source map": {
			mappings: nil,
			goLine:   10,
			goCol:    5,
			want:     nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			got := sm.FindMappingForGoPosition(tt.goLine, tt.goCol)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want %+v", tt.want)
			}
			if *got != *tt.want {
				t.Errorf("got %+v, want %+v", *got, *tt.want)
			}
		})
	}
}

func TestVirtualFileCacheAll(t *testing.T) {
	c := NewVirtualFileCache()

	if got := c.All(); len(got) != 0 {
		t.Errorf("All() on empty cache = %v, want empty", got)
	}

	smA := NewSourceMap()
	c.Put("file:///a.gsx", "file:///a_gsx_generated.go", "package a", smA, 1)
	c.Put("file:///b.gsx", "file:///b_gsx_generated.go", "package b", NewSourceMap(), 2)

	all := c.All()
	if len(all) != 2 {
		t.Fatalf("All() returned %d files, want 2", len(all))
	}

	byTui := make(map[string]*CachedVirtualFile, len(all))
	for _, f := range all {
		byTui[f.TuiURI] = f
	}
	a := byTui["file:///a.gsx"]
	if a == nil {
		t.Fatal("All() missing file:///a.gsx")
	}
	if a.GoURI != "file:///a_gsx_generated.go" || a.Content != "package a" || a.Version != 1 || a.SourceMap != smA {
		t.Errorf("cached file a = %+v, want original fields preserved", a)
	}
	if byTui["file:///b.gsx"] == nil {
		t.Error("All() missing file:///b.gsx")
	}
}
