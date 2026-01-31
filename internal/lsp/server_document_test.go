package lsp

import (
	"testing"
)

func TestDocumentManager(t *testing.T) {
	type tc struct {
		operations []func(dm *DocumentManager)
		wantDocs   int
	}

	tests := map[string]tc{
		"open single": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
			},
			wantDocs: 1,
		},
		"open multiple": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Open("file:///b.gsx", "package main", 1)
				},
			},
			wantDocs: 2,
		},
		"open and close": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Close("file:///a.gsx")
				},
			},
			wantDocs: 0,
		},
		"update": {
			operations: []func(dm *DocumentManager){
				func(dm *DocumentManager) {
					dm.Open("file:///a.gsx", "package main", 1)
				},
				func(dm *DocumentManager) {
					dm.Update("file:///a.gsx", "package updated", 2)
				},
			},
			wantDocs: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dm := NewDocumentManager()

			for _, op := range tt.operations {
				op(dm)
			}

			docs := dm.All()
			if len(docs) != tt.wantDocs {
				t.Errorf("got %d documents, want %d", len(docs), tt.wantDocs)
			}
		})
	}
}

func TestPositionConversion(t *testing.T) {
	type tc struct {
		content  string
		pos      Position
		wantOff  int
		wantBack Position
	}

	tests := map[string]tc{
		"start of file": {
			content:  "hello\nworld",
			pos:      Position{Line: 0, Character: 0},
			wantOff:  0,
			wantBack: Position{Line: 0, Character: 0},
		},
		"middle of first line": {
			content:  "hello\nworld",
			pos:      Position{Line: 0, Character: 3},
			wantOff:  3,
			wantBack: Position{Line: 0, Character: 3},
		},
		"start of second line": {
			content:  "hello\nworld",
			pos:      Position{Line: 1, Character: 0},
			wantOff:  6,
			wantBack: Position{Line: 1, Character: 0},
		},
		"middle of second line": {
			content:  "hello\nworld",
			pos:      Position{Line: 1, Character: 2},
			wantOff:  8,
			wantBack: Position{Line: 1, Character: 2},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			offset := PositionToOffset(tt.content, tt.pos)
			if offset != tt.wantOff {
				t.Errorf("PositionToOffset = %d, want %d", offset, tt.wantOff)
			}

			back := OffsetToPosition(tt.content, offset)
			if back != tt.wantBack {
				t.Errorf("OffsetToPosition = %+v, want %+v", back, tt.wantBack)
			}
		})
	}
}
