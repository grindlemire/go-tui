package tui

import "testing"

// tableWrapCellText is 29 chars: one line at its intrinsic width, wrapping
// once table columns are shrunk below it.
const tableWrapCellText = "aaaaa bbbbb ccccc ddddd eeeee"

func newWrapTestTable() (table, wideCell *Element) {
	idCell := New(WithTag("td"), WithText("id"))
	wideCell = New(WithTag("td"), WithText(tableWrapCellText))
	row := New(WithTag("tr"))
	row.AddChild(idCell)
	row.AddChild(wideCell)
	table = New(WithTag("table"))
	table.AddChild(row)
	return table, wideCell
}

// Issue #127: table columns are shrunk to fit the container, but row heights
// were computed from unwrapped intrinsic cell sizes, clipping wrapped text.
func TestTableRowHeightGrowsForShrunkWrappedCell(t *testing.T) {
	table, wideCell := newWrapTestTable()
	sentinel := New(WithText("SENTINEL"))

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(20))
	root.AddChild(table)
	root.AddChild(sentinel)

	root.Calculate(20, 24)

	cellW := wideCell.Rect().Width
	if cellW >= stringWidth(tableWrapCellText) {
		t.Fatalf("test setup: cell width %d must force wrapping", cellW)
	}
	wantLines := len(wrapText(tableWrapCellText, cellW))
	if wantLines < 2 {
		t.Fatalf("test setup: want >= 2 wrapped lines, got %d", wantLines)
	}

	if got := wideCell.Rect().Height; got != wantLines {
		t.Errorf("cell height = %d, want %d (wrapped lines)", got, wantLines)
	}
	if got := table.Rect().Height; got != wantLines {
		t.Errorf("table height = %d, want %d", got, wantLines)
	}
	if got := sentinel.Rect().Y; got != wantLines {
		t.Errorf("sentinel Y = %d, want %d (below the table)", got, wantLines)
	}
}

// HeightForWidth on a table must account for column shrinking and cell
// wrapping instead of falling into the generic row-container branch.
func TestTableHeightForWidthWrapAware(t *testing.T) {
	table, _ := newWrapTestTable()

	// At full intrinsic width nothing wraps.
	intrW, intrH := table.IntrinsicSize()
	if got := table.HeightForWidth(intrW); got != intrH {
		t.Errorf("HeightForWidth(%d) = %d, want intrinsic %d", intrW, got, intrH)
	}

	// Narrow width: columns shrink, the wide cell wraps. Column math: id
	// column keeps 2, the auto wide column absorbs the overflow.
	got := table.HeightForWidth(20)
	if got < 2 {
		t.Errorf("HeightForWidth(20) = %d, want >= 2 (wrapped cell)", got)
	}
}

// An explicit row height still wins over wrapped cell growth.
func TestTableExplicitRowHeightWinsOverWrap(t *testing.T) {
	idCell := New(WithTag("td"), WithText("id"))
	wideCell := New(WithTag("td"), WithText(tableWrapCellText))
	row := New(WithTag("tr"), WithHeight(1))
	row.AddChild(idCell)
	row.AddChild(wideCell)
	table := New(WithTag("table"))
	table.AddChild(row)

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(20))
	root.AddChild(table)

	root.Calculate(20, 24)

	if got := wideCell.Rect().Height; got != 1 {
		t.Errorf("cell height = %d, want explicit 1", got)
	}
}
