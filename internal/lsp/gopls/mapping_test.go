package gopls

import (
	"testing"
)

func TestSourceMap_RoundTrip(t *testing.T) {
	type tc struct {
		mapping  Mapping
		tuiLine  int
		tuiCol   int
		wantLine int
		wantCol  int
	}

	tests := map[string]tc{
		"exact start of mapping": {
			mapping:  Mapping{TuiLine: 5, TuiCol: 10, GoLine: 20, GoCol: 4, Length: 8},
			tuiLine:  5,
			tuiCol:   10,
			wantLine: 5,
			wantCol:  10,
		},
		"middle of mapping": {
			mapping:  Mapping{TuiLine: 5, TuiCol: 10, GoLine: 20, GoCol: 4, Length: 8},
			tuiLine:  5,
			tuiCol:   14,
			wantLine: 5,
			wantCol:  14,
		},
		"end of mapping": {
			mapping:  Mapping{TuiLine: 5, TuiCol: 10, GoLine: 20, GoCol: 4, Length: 8},
			tuiLine:  5,
			tuiCol:   18,
			wantLine: 5,
			wantCol:  18,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			sm.AddMapping(tt.mapping)

			// TUI -> Go -> TUI round trip
			goLine, goCol, found := sm.TuiToGo(tt.tuiLine, tt.tuiCol)
			if !found {
				t.Fatalf("TuiToGo(%d, %d) not found", tt.tuiLine, tt.tuiCol)
			}

			tuiLine, tuiCol, found := sm.GoToTui(goLine, goCol)
			if !found {
				t.Fatalf("GoToTui(%d, %d) not found", goLine, goCol)
			}

			if tuiLine != tt.wantLine || tuiCol != tt.wantCol {
				t.Errorf("round-trip: got (%d, %d), want (%d, %d)", tuiLine, tuiCol, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestSourceMap_TuiToGo(t *testing.T) {
	type tc struct {
		mappings []Mapping
		tuiLine  int
		tuiCol   int
		wantLine int
		wantCol  int
		wantOk   bool
	}

	tests := map[string]tc{
		"position within mapping": {
			mappings: []Mapping{{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6}},
			tuiLine:  3,
			tuiCol:   8,
			wantLine: 10,
			wantCol:  5,
			wantOk:   true,
		},
		"position outside mapping": {
			mappings: []Mapping{{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6}},
			tuiLine:  3,
			tuiCol:   0,
			wantLine: 3,
			wantCol:  0,
			wantOk:   false,
		},
		"wrong line": {
			mappings: []Mapping{{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6}},
			tuiLine:  4,
			tuiCol:   5,
			wantLine: 4,
			wantCol:  5,
			wantOk:   false,
		},
		"multiple mappings": {
			mappings: []Mapping{
				{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6},
				{TuiLine: 5, TuiCol: 10, GoLine: 15, GoCol: 8, Length: 4},
			},
			tuiLine:  5,
			tuiCol:   12,
			wantLine: 15,
			wantCol:  10,
			wantOk:   true,
		},
		"empty source map": {
			mappings: nil,
			tuiLine:  0,
			tuiCol:   0,
			wantLine: 0,
			wantCol:  0,
			wantOk:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			goLine, goCol, found := sm.TuiToGo(tt.tuiLine, tt.tuiCol)
			if found != tt.wantOk {
				t.Fatalf("TuiToGo found = %v, want %v", found, tt.wantOk)
			}
			if goLine != tt.wantLine || goCol != tt.wantCol {
				t.Errorf("TuiToGo = (%d, %d), want (%d, %d)", goLine, goCol, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestSourceMap_GoToTui(t *testing.T) {
	type tc struct {
		mappings []Mapping
		goLine   int
		goCol    int
		wantLine int
		wantCol  int
		wantOk   bool
	}

	tests := map[string]tc{
		"position within mapping": {
			mappings: []Mapping{{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6}},
			goLine:   10,
			goCol:    4,
			wantLine: 3,
			wantCol:  7,
			wantOk:   true,
		},
		"position outside mapping": {
			mappings: []Mapping{{TuiLine: 3, TuiCol: 5, GoLine: 10, GoCol: 2, Length: 6}},
			goLine:   10,
			goCol:    0,
			wantLine: 10,
			wantCol:  0,
			wantOk:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sm := NewSourceMap()
			for _, m := range tt.mappings {
				sm.AddMapping(m)
			}

			tuiLine, tuiCol, found := sm.GoToTui(tt.goLine, tt.goCol)
			if found != tt.wantOk {
				t.Fatalf("GoToTui found = %v, want %v", found, tt.wantOk)
			}
			if tuiLine != tt.wantLine || tuiCol != tt.wantCol {
				t.Errorf("GoToTui = (%d, %d), want (%d, %d)", tuiLine, tuiCol, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestSourceMap_Operations(t *testing.T) {
	sm := NewSourceMap()

	if sm.Len() != 0 {
		t.Errorf("new source map Len() = %d, want 0", sm.Len())
	}

	sm.AddMapping(Mapping{TuiLine: 1, TuiCol: 0, GoLine: 5, GoCol: 0, Length: 10})
	sm.AddMapping(Mapping{TuiLine: 2, TuiCol: 0, GoLine: 6, GoCol: 0, Length: 8})

	if sm.Len() != 2 {
		t.Errorf("after adding 2 mappings, Len() = %d, want 2", sm.Len())
	}

	if !sm.IsInGoExpression(1, 5) {
		t.Error("IsInGoExpression(1, 5) should be true")
	}

	if sm.IsInGoExpression(3, 0) {
		t.Error("IsInGoExpression(3, 0) should be false")
	}

	sm.Clear()
	if sm.Len() != 0 {
		t.Errorf("after Clear(), Len() = %d, want 0", sm.Len())
	}
}
