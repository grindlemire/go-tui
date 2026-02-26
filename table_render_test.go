package tui

import "testing"

func TestTableRender(t *testing.T) {
	type tc struct {
		buildTable  func() *Element
		width       int
		height      int
		expectCells map[[2]int]rune // (x,y) -> expected rune
		expectBold  map[[2]int]bool // (x,y) -> expected bold
	}

	tests := map[string]tc{
		"cells align across rows": {
			buildTable: func() *Element {
				table := New(WithTag("table"))
				row1 := New(WithTag("tr"))
				row1.AddChild(New(WithTag("td"), WithText("Hi")))
				row1.AddChild(New(WithTag("td"), WithText("World")))
				table.AddChild(row1)

				row2 := New(WithTag("tr"))
				row2.AddChild(New(WithTag("td"), WithText("Hello")))
				row2.AddChild(New(WithTag("td"), WithText("Go")))
				table.AddChild(row2)

				return table
			},
			width:  80,
			height: 24,
			// Col 0 width = max("Hi"=2, "Hello"=5) = 5
			// Col 1 starts at x=6 (5 + 1 gap)
			// Row 0: "Hi" at x=0, "World" at x=6
			// Row 1: "Hello" at x=0, "Go" at x=6
			expectCells: map[[2]int]rune{
				{0, 0}: 'H', {1, 0}: 'i',
				{6, 0}: 'W', {7, 0}: 'o', {8, 0}: 'r', {9, 0}: 'l', {10, 0}: 'd',
				{0, 1}: 'H', {1, 1}: 'e', {2, 1}: 'l', {3, 1}: 'l', {4, 1}: 'o',
				{6, 1}: 'G', {7, 1}: 'o',
			},
		},
		"th renders bold": {
			buildTable: func() *Element {
				table := New(WithTag("table"))
				row := New(WithTag("tr"))
				row.AddChild(New(WithTag("th"), WithText("Name")))
				table.AddChild(row)
				return table
			},
			width:  80,
			height: 24,
			expectBold: map[[2]int]bool{
				{0, 0}: true, {1, 0}: true, {2, 0}: true, {3, 0}: true,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table := tt.buildTable()
			buf := NewBuffer(tt.width, tt.height)
			table.Render(buf, tt.width, tt.height)

			for pos, expectedRune := range tt.expectCells {
				cell := buf.Cell(pos[0], pos[1])
				if cell.Rune != expectedRune {
					t.Errorf("at (%d,%d): expected rune %q, got %q",
						pos[0], pos[1], string(expectedRune), string(cell.Rune))
				}
			}

			for pos, expectedBold := range tt.expectBold {
				cell := buf.Cell(pos[0], pos[1])
				isBold := cell.Style.Attrs&AttrBold != 0
				if isBold != expectedBold {
					t.Errorf("at (%d,%d): expected bold=%v, got bold=%v",
						pos[0], pos[1], expectedBold, isBold)
				}
			}
		})
	}
}
