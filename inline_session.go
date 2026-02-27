package tui

import "strings"

func normalizeHistoryCapacity(historyCapacity int) int {
	if historyCapacity < 0 {
		return 0
	}
	return historyCapacity
}

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
	layout := inlineLayoutState{}
	layout.resetEmpty(historyCapacity)
	return layout
}

func (l *inlineLayoutState) isZeroValue() bool {
	return !l.valid && l.historyCapacity == 0 && l.contentStartRow == 0 && l.visibleRows == 0
}

func (l *inlineLayoutState) resetEmpty(historyCapacity int) {
	historyCapacity = normalizeHistoryCapacity(historyCapacity)
	l.historyCapacity = historyCapacity
	l.contentStartRow = historyCapacity
	l.visibleRows = 0
	l.valid = true
}

func (l *inlineLayoutState) resetConservativeFull(historyCapacity int) {
	historyCapacity = normalizeHistoryCapacity(historyCapacity)
	l.historyCapacity = historyCapacity
	l.contentStartRow = 0
	l.visibleRows = historyCapacity
	l.valid = true
}

func (l *inlineLayoutState) invalidate(historyCapacity int) {
	historyCapacity = normalizeHistoryCapacity(historyCapacity)
	l.historyCapacity = historyCapacity
	l.contentStartRow = 0
	l.visibleRows = 0
	l.valid = false
}

func (l *inlineLayoutState) clamp(historyCapacity int) {
	historyCapacity = normalizeHistoryCapacity(historyCapacity)
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

	// Partial line state for streaming writes.
	partialLine    []byte // accumulated bytes for current incomplete line
	partialCol     int    // visual column position (ANSI escapes excluded)
	partialVisible bool   // whether a partial line is currently displayed on screen
}

func newInlineSession(term Terminal) *inlineSession {
	return &inlineSession{terminal: term}
}

func (s *inlineSession) ensureInitialized(layout *inlineLayoutState, historyCapacity int) {
	// Zero-value layout from direct struct construction in tests/apps.
	if layout.isZeroValue() {
		*layout = newInlineLayoutState(historyCapacity)
		return
	}
	layout.clamp(historyCapacity)
}

func (s *inlineSession) invalidateForWidth(layout *inlineLayoutState, historyCapacity int) {
	layout.invalidate(historyCapacity)
}

func (s *inlineSession) appendText(layout *inlineLayoutState, historyCapacity, width int, content string) {
	if historyCapacity < 1 {
		layout.resetEmpty(historyCapacity)
		return
	}

	// After conservative invalidation, preserve existing screen by treating history
	// as full until enough appends establish a new deterministic model.
	if !layout.valid {
		layout.resetConservativeFull(historyCapacity)
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

// appendStyledText is like appendText but preserves ANSI escape sequences.
func (s *inlineSession) appendStyledText(layout *inlineLayoutState, historyCapacity, width int, content string) {
	if historyCapacity < 1 {
		layout.resetEmpty(historyCapacity)
		return
	}

	if !layout.valid {
		layout.resetConservativeFull(historyCapacity)
	}
	layout.clamp(historyCapacity)

	text := sanitizeStyledText(content)
	text = strings.TrimSuffix(text, "\n")
	rows := wrapInlineStyledRows(text, width)
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

// appendBytes processes raw bytes for streaming output. ANSI escape sequences
// are preserved. Printable characters advance the partial line. Newlines finalize
// the current partial line as a permanent row. Lines that reach terminal width
// are auto-wrapped.
func (s *inlineSession) appendBytes(layout *inlineLayoutState, historyCapacity, width int, data []byte) {
	if historyCapacity < 1 || width < 1 {
		return
	}
	if !layout.valid {
		layout.resetConservativeFull(historyCapacity)
	}
	layout.clamp(historyCapacity)

	var scanner styledByteScanner
	scanner.reset(data)

	for scanner.next() {
		switch scanner.kind {
		case tokenNewline:
			s.commitPartialRow(layout)
		case tokenANSI:
			s.partialLine = append(s.partialLine, scanner.bytes()...)
		case tokenRune:
			w := scanner.runeWidth
			// Wrap if this rune would exceed width.
			if s.partialCol+w > width {
				s.commitPartialRow(layout)
			}
			s.partialLine = append(s.partialLine, scanner.bytes()...)
			s.partialCol += w
		}
	}

	// Display the current partial line in-place.
	s.displayPartial(layout, historyCapacity)
}

// commitPartialRow finalizes the current partial line as a permanent history row.
func (s *inlineSession) commitPartialRow(layout *inlineLayoutState) {
	if !s.partialVisible {
		// Not yet displayed — commit to layout via appendRow.
		var seq strings.Builder
		s.appendRow(&seq, layout, string(s.partialLine))
		if seq.Len() > 0 {
			s.terminal.WriteDirect([]byte(seq.String()))
		}
	}
	// else: already displayed via displayPartial, row is in the layout.
	s.partialLine = s.partialLine[:0]
	s.partialCol = 0
	s.partialVisible = false
}

// displayPartial writes the current partial line to the terminal. On first
// call it claims a layout slot via appendRow; subsequent calls overwrite in-place.
func (s *inlineSession) displayPartial(layout *inlineLayoutState, historyCapacity int) {
	if len(s.partialLine) == 0 && s.partialCol == 0 {
		return
	}

	if !s.partialVisible {
		// Claim a new row slot via appendRow.
		var seq strings.Builder
		s.appendRow(&seq, layout, string(s.partialLine))
		if seq.Len() > 0 {
			s.terminal.WriteDirect([]byte(seq.String()))
		}
		s.partialVisible = true
	} else {
		// Overwrite the existing partial row in-place.
		targetRow := layout.contentStartRow + layout.visibleRows - 1
		var seq strings.Builder
		inlineAppendUpdateLine(&seq, targetRow, string(s.partialLine))
		if seq.Len() > 0 {
			s.terminal.WriteDirect([]byte(seq.String()))
		}
	}
}

// finalizePartial commits any in-progress partial line as a permanent row.
// No-op if the partial line is empty.
func (s *inlineSession) finalizePartial(layout *inlineLayoutState) {
	if s.partialCol == 0 && len(s.partialLine) == 0 {
		return
	}
	if !s.partialVisible {
		// Partial was accumulated but never displayed — commit it now.
		var seq strings.Builder
		s.appendRow(&seq, layout, string(s.partialLine))
		if seq.Len() > 0 {
			s.terminal.WriteDirect([]byte(seq.String()))
		}
	}
	s.partialLine = s.partialLine[:0]
	s.partialCol = 0
	s.partialVisible = false
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

	oldHistoryCap := normalizeHistoryCapacity(oldStartRow)
	newHistoryCap := normalizeHistoryCapacity(newStartRow)

	if !layout.valid {
		layout.clamp(newHistoryCap)
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
			// Consume only blank rows at the top of the history area; no content moves
			// to scrollback and row 0 is untouched.
			topRow := topBlanks - remaining
			inlineAppendScrollUp(&seq, topRow, historyCapacity-1, remaining)
			layout.contentStartRow -= remaining
			remaining = 0

		case topBlanks > 1:
			// Consume as many top blanks as possible while preserving row 0.
			// This avoids introducing an extra blank row into scrollback.
			consume := topBlanks - 1
			if consume > remaining {
				consume = remaining
			}
			topRow := topBlanks - consume
			inlineAppendScrollUp(&seq, topRow, historyCapacity-1, consume)
			layout.contentStartRow -= consume
			remaining -= consume

		default:
			// We have exhausted top blank slack (or only row 0 remains blank), so
			// scroll from row 0 and account for any content rows pushed away.
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
