package tui

import "strings"

// inlineLayoutState tracks the current visible history geometry above the widget.
type inlineLayoutState struct {
	// Number of history rows available above the widget.
	historyCapacity int
	// Row index (0-based) where the oldest visible history row starts.
	contentStartRow int
	// Count of visible history rows in the content block.
	visibleRows int
	// Whether the geometry is trustworthy for precise operations.
	valid bool
}

func newInlineLayoutState(historyCapacity int) inlineLayoutState {
	if historyCapacity < 0 {
		historyCapacity = 0
	}
	return inlineLayoutState{
		historyCapacity: historyCapacity,
		contentStartRow: historyCapacity,
		visibleRows:     0,
		valid:           true,
	}
}

func (l *inlineLayoutState) clamp(historyCapacity int) {
	if historyCapacity < 0 {
		historyCapacity = 0
	}
	l.historyCapacity = historyCapacity

	if !l.valid {
		return
	}

	if l.visibleRows < 0 {
		l.visibleRows = 0
	}
	if l.visibleRows > historyCapacity {
		l.visibleRows = historyCapacity
	}
	if l.visibleRows == 0 {
		l.contentStartRow = historyCapacity
		return
	}

	maxStart := historyCapacity - l.visibleRows
	if l.contentStartRow < 0 {
		l.contentStartRow = 0
	}
	if l.contentStartRow > maxStart {
		l.contentStartRow = maxStart
	}
}

type inlineSession struct {
	terminal Terminal
}

func newInlineSession(term Terminal) *inlineSession {
	return &inlineSession{terminal: term}
}

func (s *inlineSession) ensureInitialized(layout *inlineLayoutState, historyCapacity int) {
	// Zero-value layout from direct struct construction in tests/apps.
	if !layout.valid && layout.historyCapacity == 0 && layout.contentStartRow == 0 && layout.visibleRows == 0 {
		*layout = newInlineLayoutState(historyCapacity)
		return
	}

	if layout.historyCapacity != historyCapacity {
		layout.historyCapacity = historyCapacity
	}
	layout.clamp(historyCapacity)
}

func (s *inlineSession) invalidateForWidth(layout *inlineLayoutState, historyCapacity int) {
	if historyCapacity < 0 {
		historyCapacity = 0
	}
	layout.historyCapacity = historyCapacity
	layout.contentStartRow = 0
	layout.visibleRows = 0
	layout.valid = false
}

func (s *inlineSession) appendText(layout *inlineLayoutState, historyCapacity, width int, content string) {
	if historyCapacity < 1 {
		layout.historyCapacity = historyCapacity
		layout.contentStartRow = historyCapacity
		layout.visibleRows = 0
		layout.valid = true
		return
	}

	// After conservative invalidation, preserve existing screen by treating history
	// as full until enough appends establish a new deterministic model.
	if !layout.valid {
		layout.historyCapacity = historyCapacity
		layout.contentStartRow = 0
		layout.visibleRows = historyCapacity
		layout.valid = true
	}
	layout.clamp(historyCapacity)

	text := sanitizeInlineText(content)
	text = strings.TrimSuffix(text, "\n")
	rows := wrapInlineVisualRows(text, width)
	if len(rows) == 0 {
		return
	}

	var seq strings.Builder
	for _, row := range rows {
		s.appendRow(&seq, layout, row)
	}

	if seq.Len() > 0 {
		s.terminal.WriteDirect([]byte(seq.String()))
	}
}

func (s *inlineSession) appendRow(seq *strings.Builder, layout *inlineLayoutState, row string) {
	historyCapacity := layout.historyCapacity
	if historyCapacity < 1 {
		return
	}

	if layout.visibleRows == 0 {
		target := historyCapacity - 1
		inlineAppendWriteLine(seq, target, row)
		layout.contentStartRow = target
		layout.visibleRows = 1
		return
	}

	contentEndRow := layout.contentStartRow + layout.visibleRows - 1
	bottomBlanks := (historyCapacity - 1) - contentEndRow
	if bottomBlanks > 0 {
		target := contentEndRow + 1
		inlineAppendWriteLine(seq, target, row)
		layout.visibleRows++
		layout.clamp(historyCapacity)
		return
	}

	topRow := layout.contentStartRow
	if layout.visibleRows < historyCapacity && topRow > 0 {
		// Expand block upward by consuming one blank row.
		topRow--
	}

	inlineAppendScrollUp(seq, topRow, historyCapacity-1, 1)
	inlineAppendWriteLine(seq, historyCapacity-1, row)

	if layout.visibleRows < historyCapacity {
		layout.visibleRows++
		if layout.contentStartRow > 0 {
			layout.contentStartRow--
		}
	} else {
		layout.contentStartRow = 0
	}
	layout.clamp(historyCapacity)
}

func (s *inlineSession) resize(layout *inlineLayoutState, oldStartRow, oldHeight, newStartRow int) {
	s.clearWidgetArea(oldStartRow, oldHeight)

	oldHistoryCap := oldStartRow
	newHistoryCap := newStartRow
	if oldHistoryCap < 0 {
		oldHistoryCap = 0
	}
	if newHistoryCap < 0 {
		newHistoryCap = 0
	}

	if !layout.valid {
		layout.historyCapacity = newHistoryCap
		return
	}

	layout.clamp(oldHistoryCap)
	if newHistoryCap < oldHistoryCap {
		s.consumeForGrowth(layout, oldHistoryCap, oldHistoryCap-newHistoryCap)
	}

	layout.clamp(newHistoryCap)
}

func (s *inlineSession) clearWidgetArea(startRow, height int) {
	if height < 1 || startRow < 0 {
		return
	}
	var seq strings.Builder
	inlineAppendClearRows(&seq, startRow, height)
	if seq.Len() > 0 {
		s.terminal.WriteDirect([]byte(seq.String()))
	}
}

// consumeForGrowth removes rows from the history region when the widget grows.
// Rows are consumed from top blanks first; once exhausted, oldest content rows
// are scrolled into terminal scrollback.
func (s *inlineSession) consumeForGrowth(layout *inlineLayoutState, historyCapacity, lines int) {
	if historyCapacity < 1 || lines < 1 {
		return
	}
	if layout.visibleRows < 1 {
		return
	}

	remaining := lines
	var seq strings.Builder

	for remaining > 0 {
		topBlanks := layout.contentStartRow

		switch {
		case topBlanks > remaining:
			topRow := topBlanks - remaining
			inlineAppendScrollUp(&seq, topRow, historyCapacity-1, remaining)
			layout.contentStartRow -= remaining
			remaining = 0

		case topBlanks > 1:
			consume := topBlanks - 1
			if consume > remaining {
				consume = remaining
			}
			topRow := topBlanks - consume
			inlineAppendScrollUp(&seq, topRow, historyCapacity-1, consume)
			layout.contentStartRow -= consume
			remaining -= consume

		default:
			consume := remaining
			inlineAppendScrollUp(&seq, 0, historyCapacity-1, consume)

			removedContent := consume - topBlanks
			if removedContent < 0 {
				removedContent = 0
			}
			if removedContent > layout.visibleRows {
				removedContent = layout.visibleRows
			}

			layout.visibleRows -= removedContent
			layout.contentStartRow = 0
			remaining = 0
		}
	}

	if seq.Len() > 0 {
		s.terminal.WriteDirect([]byte(seq.String()))
	}
}

func (a *App) invalidateInlineLayoutForWidthChange(historyCapacity int) {
	if a.inlineHeight == 0 {
		return
	}
	a.ensureInlineSession()
	a.inlineSession.ensureInitialized(&a.inlineLayout, historyCapacity)
	a.inlineSession.invalidateForWidth(&a.inlineLayout, historyCapacity)
}
