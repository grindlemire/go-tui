package layout

import "testing"

// tableNode extends testNode with a configurable tag.
// This is needed because testNode.Tag() always returns "".
type tableNode struct {
	testNode
	tag string
}

func newTableNode(tag string, style Style) *tableNode {
	return &tableNode{
		testNode: testNode{
			style: style,
			dirty: true,
		},
		tag: tag,
	}
}

func (n *tableNode) Tag() string { return n.tag }

func (n *tableNode) LayoutChildren() []Layoutable {
	result := make([]Layoutable, len(n.children))
	for i, child := range n.children {
		result[i] = child
	}
	return result
}

// addTableChild appends a tableNode child and marks dirty.
func (n *tableNode) addTableChild(children ...*tableNode) {
	for _, child := range children {
		child.parent = &n.testNode
		n.children = append(n.children, &child.testNode)
	}
	n.markDirty()
}

// Compile-time check that tableNode implements Layoutable.
var _ Layoutable = (*tableNode)(nil)

func TestLayoutTable(t *testing.T) {
	type cellExpect struct {
		width  int
		height int
	}

	type tc struct {
		desc       string
		tableWidth int
		tableHeight int
		build      func() *tableNode
		// expectations: rows[rowIdx][colIdx] = expected cell dimensions
		cells [][]cellExpect
	}

	tests := map[string]tc{
		"basic 2x2 auto-sized columns": {
			desc:        "all cells same intrinsic width; columns should match",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row1 := newTableNode("tr", DefaultStyle())
				cell1a := newTableNode("td", DefaultStyle())
				cell1a.intrinsicW = 10
				cell1a.intrinsicH = 1
				cell1b := newTableNode("td", DefaultStyle())
				cell1b.intrinsicW = 10
				cell1b.intrinsicH = 1
				row1.addTableChild(cell1a, cell1b)

				row2 := newTableNode("tr", DefaultStyle())
				cell2a := newTableNode("td", DefaultStyle())
				cell2a.intrinsicW = 10
				cell2a.intrinsicH = 1
				cell2b := newTableNode("td", DefaultStyle())
				cell2b.intrinsicW = 10
				cell2b.intrinsicH = 1
				row2.addTableChild(cell2a, cell2b)

				table.addTableChild(row1, row2)
				return table
			},
			cells: [][]cellExpect{
				{{width: 10, height: 1}, {width: 10, height: 1}},
				{{width: 10, height: 1}, {width: 10, height: 1}},
			},
		},
		"widest cell wins per column": {
			desc:        "column widths should be the max across all rows",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row1 := newTableNode("tr", DefaultStyle())
				cell1a := newTableNode("td", DefaultStyle())
				cell1a.intrinsicW = 5
				cell1a.intrinsicH = 1
				cell1b := newTableNode("td", DefaultStyle())
				cell1b.intrinsicW = 20
				cell1b.intrinsicH = 1
				row1.addTableChild(cell1a, cell1b)

				row2 := newTableNode("tr", DefaultStyle())
				cell2a := newTableNode("td", DefaultStyle())
				cell2a.intrinsicW = 15
				cell2a.intrinsicH = 1
				cell2b := newTableNode("td", DefaultStyle())
				cell2b.intrinsicW = 8
				cell2b.intrinsicH = 1
				row2.addTableChild(cell2a, cell2b)

				table.addTableChild(row1, row2)
				return table
			},
			cells: [][]cellExpect{
				{{width: 15, height: 1}, {width: 20, height: 1}},
				{{width: 15, height: 1}, {width: 20, height: 1}},
			},
		},
		"rows with fewer cells padded": {
			desc:        "short row should still get laid out correctly",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row1 := newTableNode("tr", DefaultStyle())
				cell1a := newTableNode("td", DefaultStyle())
				cell1a.intrinsicW = 10
				cell1a.intrinsicH = 1
				cell1b := newTableNode("td", DefaultStyle())
				cell1b.intrinsicW = 10
				cell1b.intrinsicH = 1
				cell1c := newTableNode("td", DefaultStyle())
				cell1c.intrinsicW = 10
				cell1c.intrinsicH = 1
				row1.addTableChild(cell1a, cell1b, cell1c)

				// Row 2 has fewer cells
				row2 := newTableNode("tr", DefaultStyle())
				cell2a := newTableNode("td", DefaultStyle())
				cell2a.intrinsicW = 12
				cell2a.intrinsicH = 2
				row2.addTableChild(cell2a)

				table.addTableChild(row1, row2)
				return table
			},
			cells: [][]cellExpect{
				{{width: 12, height: 1}, {width: 10, height: 1}, {width: 10, height: 1}},
				{{width: 12, height: 2}}, // only one cell in this row
			},
		},
		"single row single column": {
			desc:        "edge case: 1x1 table",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row := newTableNode("tr", DefaultStyle())
				cell := newTableNode("td", DefaultStyle())
				cell.intrinsicW = 25
				cell.intrinsicH = 3
				row.addTableChild(cell)

				table.addTableChild(row)
				return table
			},
			cells: [][]cellExpect{
				{{width: 25, height: 3}},
			},
		},
		"explicit height on tr overrides computed": {
			desc:        "row with Fixed(5) height should override max cell height",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row := newTableNode("tr", DefaultStyle())
				row.style.Height = Fixed(5) // explicit row height
				cell := newTableNode("td", DefaultStyle())
				cell.intrinsicW = 10
				cell.intrinsicH = 1 // intrinsic is 1, but row forces 5
				row.addTableChild(cell)

				table.addTableChild(row)
				return table
			},
			cells: [][]cellExpect{
				{{width: 10, height: 5}},
			},
		},
		"explicit width on cell overrides auto": {
			desc:        "cell with Fixed(20) width should override intrinsic",
			tableWidth:  80,
			tableHeight: 24,
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				table.style.Width = Fixed(80)
				table.style.Height = Fixed(24)

				row1 := newTableNode("tr", DefaultStyle())
				cell1a := newTableNode("td", DefaultStyle())
				cell1a.style.Width = Fixed(20)
				cell1a.intrinsicW = 5
				cell1a.intrinsicH = 1
				cell1b := newTableNode("td", DefaultStyle())
				cell1b.intrinsicW = 10
				cell1b.intrinsicH = 1
				row1.addTableChild(cell1a, cell1b)

				row2 := newTableNode("tr", DefaultStyle())
				cell2a := newTableNode("td", DefaultStyle())
				cell2a.intrinsicW = 8
				cell2a.intrinsicH = 1
				cell2b := newTableNode("td", DefaultStyle())
				cell2b.intrinsicW = 10
				cell2b.intrinsicH = 1
				row2.addTableChild(cell2a, cell2b)

				table.addTableChild(row1, row2)
				return table
			},
			cells: [][]cellExpect{
				// Column 0: max(Fixed(20), intrinsic 8) = 20
				// Column 1: max(intrinsic 10, intrinsic 10) = 10
				{{width: 20, height: 1}, {width: 10, height: 1}},
				{{width: 20, height: 1}, {width: 10, height: 1}},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table := tt.build()
			Calculate(table, tt.tableWidth, tt.tableHeight)

			rows := table.LayoutChildren()
			if len(rows) != len(tt.cells) {
				t.Fatalf("expected %d rows, got %d", len(tt.cells), len(rows))
			}

			for ri, row := range rows {
				cells := row.LayoutChildren()
				if len(cells) != len(tt.cells[ri]) {
					t.Fatalf("row %d: expected %d cells, got %d", ri, len(tt.cells[ri]), len(cells))
				}

				for ci, cell := range cells {
					layout := cell.GetLayout()
					expect := tt.cells[ri][ci]

					if layout.Rect.Width != expect.width {
						t.Errorf("row %d, col %d: width = %d, want %d",
							ri, ci, layout.Rect.Width, expect.width)
					}
					if layout.Rect.Height != expect.height {
						t.Errorf("row %d, col %d: height = %d, want %d",
							ri, ci, layout.Rect.Height, expect.height)
					}
				}
			}
		})
	}
}

func TestLayoutTable_RowPositions(t *testing.T) {
	// Verify that rows are stacked top-to-bottom and cells are positioned left-to-right.
	table := newTableNode("table", DefaultStyle())
	table.style.Width = Fixed(80)
	table.style.Height = Fixed(24)

	row1 := newTableNode("tr", DefaultStyle())
	cell1a := newTableNode("td", DefaultStyle())
	cell1a.intrinsicW = 10
	cell1a.intrinsicH = 2
	cell1b := newTableNode("td", DefaultStyle())
	cell1b.intrinsicW = 15
	cell1b.intrinsicH = 1
	row1.addTableChild(cell1a, cell1b)

	row2 := newTableNode("tr", DefaultStyle())
	cell2a := newTableNode("td", DefaultStyle())
	cell2a.intrinsicW = 10
	cell2a.intrinsicH = 3
	cell2b := newTableNode("td", DefaultStyle())
	cell2b.intrinsicW = 15
	cell2b.intrinsicH = 1
	row2.addTableChild(cell2a, cell2b)

	table.addTableChild(row1, row2)
	Calculate(table, 80, 24)

	// Row 1 should start at Y=0, Row 2 at Y=2 (row 1 height = max(2,1) = 2)
	row1Layout := row1.GetLayout()
	row2Layout := row2.GetLayout()

	if row1Layout.Rect.Y != 0 {
		t.Errorf("row1 Y = %d, want 0", row1Layout.Rect.Y)
	}
	if row1Layout.Rect.Height != 2 {
		t.Errorf("row1 Height = %d, want 2", row1Layout.Rect.Height)
	}
	if row2Layout.Rect.Y != 2 {
		t.Errorf("row2 Y = %d, want 2", row2Layout.Rect.Y)
	}
	if row2Layout.Rect.Height != 3 {
		t.Errorf("row2 Height = %d, want 3", row2Layout.Rect.Height)
	}

	// Cell positions: cell1a at X=0, cell1b at X=10
	cells1 := row1.LayoutChildren()
	cell1aLayout := cells1[0].GetLayout()
	cell1bLayout := cells1[1].GetLayout()

	if cell1aLayout.Rect.X != 0 {
		t.Errorf("cell1a X = %d, want 0", cell1aLayout.Rect.X)
	}
	if cell1bLayout.Rect.X != 10 {
		t.Errorf("cell1b X = %d, want 10", cell1bLayout.Rect.X)
	}

	// Cell heights should match their row's height
	if cell1aLayout.Rect.Height != 2 {
		t.Errorf("cell1a Height = %d, want 2 (row height)", cell1aLayout.Rect.Height)
	}
	if cell1bLayout.Rect.Height != 2 {
		t.Errorf("cell1b Height = %d, want 2 (row height)", cell1bLayout.Rect.Height)
	}
}

func TestLayoutTable_ShrinkColumns(t *testing.T) {
	// When total column widths exceed available width, auto columns should shrink.
	table := newTableNode("table", DefaultStyle())
	table.style.Width = Fixed(30)
	table.style.Height = Fixed(10)

	row := newTableNode("tr", DefaultStyle())
	cell1 := newTableNode("td", DefaultStyle())
	cell1.intrinsicW = 20
	cell1.intrinsicH = 1
	cell2 := newTableNode("td", DefaultStyle())
	cell2.intrinsicW = 20
	cell2.intrinsicH = 1
	row.addTableChild(cell1, cell2)

	table.addTableChild(row)
	Calculate(table, 30, 10)

	cells := row.LayoutChildren()
	c1Layout := cells[0].GetLayout()
	c2Layout := cells[1].GetLayout()

	// Total intrinsic = 40, available = 30
	// Both are auto so they shrink proportionally: each gets 15
	if c1Layout.Rect.Width != 15 {
		t.Errorf("cell1 width = %d, want 15 (shrunk)", c1Layout.Rect.Width)
	}
	if c2Layout.Rect.Width != 15 {
		t.Errorf("cell2 width = %d, want 15 (shrunk)", c2Layout.Rect.Width)
	}
}

func TestTableIntrinsicSize(t *testing.T) {
	type tc struct {
		build       func() *tableNode
		expectWidth int
		expectHeight int
	}

	tests := map[string]tc{
		"basic 2x2": {
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())

				row1 := newTableNode("tr", DefaultStyle())
				c1a := newTableNode("td", DefaultStyle())
				c1a.intrinsicW = 10
				c1a.intrinsicH = 1
				c1b := newTableNode("td", DefaultStyle())
				c1b.intrinsicW = 20
				c1b.intrinsicH = 1
				row1.addTableChild(c1a, c1b)

				row2 := newTableNode("tr", DefaultStyle())
				c2a := newTableNode("td", DefaultStyle())
				c2a.intrinsicW = 15
				c2a.intrinsicH = 2
				c2b := newTableNode("td", DefaultStyle())
				c2b.intrinsicW = 8
				c2b.intrinsicH = 1
				row2.addTableChild(c2a, c2b)

				table.addTableChild(row1, row2)
				return table
			},
			// Col 0: max(10, 15) = 15, Col 1: max(20, 8) = 20 => width = 35
			// Row 0: max(1, 1) = 1, Row 1: max(2, 1) = 2 => height = 3
			expectWidth:  35,
			expectHeight: 3,
		},
		"empty table": {
			build: func() *tableNode {
				return newTableNode("table", DefaultStyle())
			},
			expectWidth:  0,
			expectHeight: 0,
		},
		"single cell": {
			build: func() *tableNode {
				table := newTableNode("table", DefaultStyle())
				row := newTableNode("tr", DefaultStyle())
				cell := newTableNode("td", DefaultStyle())
				cell.intrinsicW = 12
				cell.intrinsicH = 4
				row.addTableChild(cell)
				table.addTableChild(row)
				return table
			},
			expectWidth:  12,
			expectHeight: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table := tt.build()
			w, h := TableIntrinsicSize(table)

			if w != tt.expectWidth {
				t.Errorf("intrinsic width = %d, want %d", w, tt.expectWidth)
			}
			if h != tt.expectHeight {
				t.Errorf("intrinsic height = %d, want %d", h, tt.expectHeight)
			}
		})
	}
}
