package tui

import "testing"

// rowWrapLongText is 12 five-char words (71 chars): 1 line at its intrinsic
// width, 2 lines at width 40, 3 lines at width 30.
const rowWrapLongText = "aaaaa bbbbb ccccc ddddd eeeee fffff ggggg hhhhh iiiii jjjjj kkkkk lllll"

// Reproduction for issue #126: a wrapping text child in a horizontal flex row
// receives its final width from flex distribution, but the row's height is
// estimated from pre-flex (full content width) wrapping, so the row does not
// grow when the post-flex width wraps to more lines.
func TestRowHeightGrowsForPostFlexWrappedChild(t *testing.T) {
	fullLines := len(wrapText(rowWrapLongText, 40))
	flexLines := len(wrapText(rowWrapLongText, 30))
	if flexLines <= fullLines {
		t.Fatalf("test setup: wrap at 30 (%d lines) must exceed wrap at 40 (%d lines)", flexLines, fullLines)
	}

	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0), WithWrap(false))
	wrapped := New(WithText(rowWrapLongText), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100))
	row.AddChild(label)
	row.AddChild(wrapped)

	sentinel := New(WithText("SENTINEL"))

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)
	root.AddChild(sentinel)

	root.Calculate(40, 24)

	if got := wrapped.Rect().Width; got != 30 {
		t.Fatalf("wrapped child width = %d, want 30", got)
	}
	if got := row.Rect().Height; got != flexLines {
		t.Errorf("row height = %d, want %d (post-flex wrapped lines)", got, flexLines)
	}
	if got := wrapped.Rect().Height; got != flexLines {
		t.Errorf("wrapped child height = %d, want %d", got, flexLines)
	}
	if got := sentinel.Rect().Y; got != flexLines {
		t.Errorf("sentinel Y = %d, want %d (below the wrapped row)", got, flexLines)
	}
}

// Same bug class for flex-wrap rows: an item pushed to its own wrap line is
// shrunk to the row width, wrapping its text to more lines than the pre-flex
// base-size estimate predicted.
func TestWrapRowHeightGrowsForPostFlexWrappedChild(t *testing.T) {
	lineLines := len(wrapText(rowWrapLongText, 40))
	if lineLines != 2 {
		t.Fatalf("test setup: wrap at 40 = %d lines, want 2", lineLines)
	}

	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0), WithWrap(false))
	wrapped := New(WithText(rowWrapLongText), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100), WithFlexWrap(Wrap))
	row.AddChild(label)
	row.AddChild(wrapped)

	sentinel := New(WithText("SENTINEL"))

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)
	root.AddChild(sentinel)

	root.Calculate(40, 24)

	// label occupies wrap line 1 (1 row); the shrunk text child occupies
	// wrap line 2 (2 rows).
	wantRowH := 1 + lineLines
	if got := wrapped.Rect().Width; got != 40 {
		t.Fatalf("wrapped child width = %d, want 40", got)
	}
	if got := row.Rect().Height; got != wantRowH {
		t.Errorf("row height = %d, want %d (sum of wrap line heights)", got, wantRowH)
	}
	if got := sentinel.Rect().Y; got != wantRowH {
		t.Errorf("sentinel Y = %d, want %d (below the wrapped row)", got, wantRowH)
	}
}

// The post-flex width estimate must account for the row's gap.
func TestRowHeightForWidthAccountsForGap(t *testing.T) {
	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0), WithWrap(false))
	wrapped := New(WithText(rowWrapLongText), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100), WithGap(2))
	row.AddChild(label)
	row.AddChild(wrapped)

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)

	root.Calculate(40, 24)

	wantLines := len(wrapText(rowWrapLongText, 28)) // 40 - 10 label - 2 gap
	if got := wrapped.Rect().Width; got != 28 {
		t.Fatalf("wrapped child width = %d, want 28", got)
	}
	if got := row.Rect().Height; got != wantLines {
		t.Errorf("row height = %d, want %d", got, wantLines)
	}
}

// A row whose text fits without wrapping keeps its single-line height.
func TestRowHeightUnchangedWhenNothingWraps(t *testing.T) {
	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0), WithWrap(false))
	short := New(WithText("short"), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100))
	row.AddChild(label)
	row.AddChild(short)

	sentinel := New(WithText("SENTINEL"))

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)
	root.AddChild(sentinel)

	root.Calculate(40, 24)

	if got := row.Rect().Height; got != 1 {
		t.Errorf("row height = %d, want 1", got)
	}
	if got := sentinel.Rect().Y; got != 1 {
		t.Errorf("sentinel Y = %d, want 1", got)
	}
}

// An explicit row height wins over the post-flex wrapped estimate.
func TestRowExplicitHeightWinsOverWrappedEstimate(t *testing.T) {
	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0), WithWrap(false))
	wrapped := New(WithText(rowWrapLongText), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100), WithHeight(2))
	row.AddChild(label)
	row.AddChild(wrapped)

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)

	root.Calculate(40, 24)

	if got := row.HeightForWidth(40); got != 2 {
		t.Errorf("HeightForWidth(40) = %d, want explicit 2", got)
	}
	if got := row.Rect().Height; got != 2 {
		t.Errorf("row height = %d, want explicit 2", got)
	}
}

// A child's explicit height must win over its intrinsic height in the row
// estimate: a bordered text child clamped to height 1 has intrinsic height 3
// (1 line + 2 border rows), but the row must not grow past the clamp.
func TestRowHeightHonorsExplicitChildHeight(t *testing.T) {
	clamped := New(WithText("x"), WithBorder(BorderSingle), WithHeight(1))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row), WithWidthPercent(100))
	row.AddChild(clamped)

	sentinel := New(WithText("SENTINEL"))

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(40))
	root.AddChild(row)
	root.AddChild(sentinel)

	root.Calculate(40, 24)

	if got := row.HeightForWidth(40); got != 1 {
		t.Errorf("row HeightForWidth(40) = %d, want 1 (explicit child height)", got)
	}
	if got := row.Rect().Height; got != 1 {
		t.Errorf("row height = %d, want 1", got)
	}
	if got := sentinel.Rect().Y; got != 1 {
		t.Errorf("sentinel Y = %d, want 1", got)
	}
}

// Degenerate widths must not panic and must not report negative heights.
func TestRowHeightForWidthDegenerateWidth(t *testing.T) {
	label := New(WithText("label"), WithWidth(10), WithFlexShrink(0))
	wrapped := New(WithText("some wrapping text here"), WithFlexGrow(1), WithMinWidth(0))

	row := New(WithDisplay(DisplayFlex), WithDirection(Row))
	row.AddChild(label)
	row.AddChild(wrapped)

	root := New(WithDisplay(DisplayFlex), WithDirection(Column), WithWidth(0))
	root.AddChild(row)

	root.Calculate(0, 24)

	if got := row.HeightForWidth(0); got < 0 {
		t.Errorf("HeightForWidth(0) = %d, want >= 0", got)
	}
}
