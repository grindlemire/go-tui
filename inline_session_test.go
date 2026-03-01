package tui

import (
	"testing"
)

func TestInlineLayoutState_NewEmpty(t *testing.T) {
	type tc struct {
		historyCapacity int
		wantStartRow    int
		wantVisible     int
		wantValid       bool
	}

	tests := map[string]tc{
		"positive capacity": {
			historyCapacity: 10,
			wantStartRow:    10,
			wantVisible:     0,
			wantValid:       true,
		},
		"zero capacity": {
			historyCapacity: 0,
			wantStartRow:    0,
			wantVisible:     0,
			wantValid:       true,
		},
		"negative capacity normalized to zero": {
			historyCapacity: -5,
			wantStartRow:    0,
			wantVisible:     0,
			wantValid:       true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			layout := newInlineLayoutState(tt.historyCapacity)
			if layout.contentStartRow != tt.wantStartRow {
				t.Errorf("contentStartRow = %d, want %d", layout.contentStartRow, tt.wantStartRow)
			}
			if layout.visibleRows != tt.wantVisible {
				t.Errorf("visibleRows = %d, want %d", layout.visibleRows, tt.wantVisible)
			}
			if layout.valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", layout.valid, tt.wantValid)
			}
		})
	}
}

func TestInlineLayoutState_Invalidate(t *testing.T) {
	layout := newInlineLayoutState(10)
	if !layout.valid {
		t.Fatal("new layout should be valid")
	}

	layout.invalidate(10)
	if layout.valid {
		t.Error("layout should be invalid after invalidate()")
	}
	if layout.visibleRows != 0 {
		t.Errorf("visibleRows = %d, want 0 after invalidate", layout.visibleRows)
	}
}

func TestInlineLayoutState_ResetConservativeFull(t *testing.T) {
	layout := newInlineLayoutState(10)
	layout.invalidate(10)

	layout.resetConservativeFull(10)
	if !layout.valid {
		t.Error("should be valid after resetConservativeFull")
	}
	if layout.contentStartRow != 0 {
		t.Errorf("contentStartRow = %d, want 0", layout.contentStartRow)
	}
	if layout.visibleRows != 10 {
		t.Errorf("visibleRows = %d, want 10", layout.visibleRows)
	}
}

func TestInlineLayoutState_Clamp(t *testing.T) {
	type tc struct {
		setup           func() inlineLayoutState
		historyCapacity int
		wantStartRow    int
		wantVisible     int
	}

	tests := map[string]tc{
		"clamp visible rows to capacity": {
			setup: func() inlineLayoutState {
				l := newInlineLayoutState(20)
				l.visibleRows = 15
				l.contentStartRow = 0
				return l
			},
			historyCapacity: 10,
			wantStartRow:    0,
			wantVisible:     10,
		},
		"clamp negative visible rows": {
			setup: func() inlineLayoutState {
				l := newInlineLayoutState(10)
				l.visibleRows = -3
				return l
			},
			historyCapacity: 10,
			wantStartRow:    10,
			wantVisible:     0,
		},
		"clamp start row to valid range": {
			setup: func() inlineLayoutState {
				l := newInlineLayoutState(10)
				l.visibleRows = 5
				l.contentStartRow = 8 // max is 10-5=5
				return l
			},
			historyCapacity: 10,
			wantStartRow:    5,
			wantVisible:     5,
		},
		"invalid layout stays invalid": {
			setup: func() inlineLayoutState {
				l := newInlineLayoutState(10)
				l.invalidate(10)
				return l
			},
			historyCapacity: 10,
			wantStartRow:    0,
			wantVisible:     0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			layout := tt.setup()
			layout.clamp(tt.historyCapacity)
			if layout.visibleRows != tt.wantVisible {
				t.Errorf("visibleRows = %d, want %d", layout.visibleRows, tt.wantVisible)
			}
			if layout.contentStartRow != tt.wantStartRow {
				t.Errorf("contentStartRow = %d, want %d", layout.contentStartRow, tt.wantStartRow)
			}
		})
	}
}

func TestInlineLayoutState_IsZeroValue(t *testing.T) {
	type tc struct {
		layout inlineLayoutState
		want   bool
	}

	tests := map[string]tc{
		"zero value struct": {
			layout: inlineLayoutState{},
			want:   true,
		},
		"initialized layout": {
			layout: newInlineLayoutState(10),
			want:   false,
		},
		"valid but otherwise zero": {
			layout: inlineLayoutState{valid: true},
			want:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.layout.isZeroValue(); got != tt.want {
				t.Errorf("isZeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeHistoryCapacity(t *testing.T) {
	type tc struct {
		input int
		want  int
	}

	tests := map[string]tc{
		"positive":  {input: 10, want: 10},
		"zero":      {input: 0, want: 0},
		"negative":  {input: -5, want: 0},
		"large neg": {input: -100, want: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := normalizeHistoryCapacity(tt.input); got != tt.want {
				t.Errorf("normalizeHistoryCapacity(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
