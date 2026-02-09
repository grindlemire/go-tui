package tui

import (
	"fmt"
	"strings"
)

func inlineAppendWriteLine(seq *strings.Builder, row int, text string) {
	if row < 0 {
		return
	}
	seq.WriteString(fmt.Sprintf("\033[%d;1H\033[2K", row+1))
	seq.WriteString(text)
}

func inlineAppendScrollUp(seq *strings.Builder, topRow, bottomRow, n int) {
	if n < 1 || topRow < 0 || bottomRow < 0 || topRow >= bottomRow+1 {
		return
	}
	seq.WriteString(fmt.Sprintf("\033[%d;%dr", topRow+1, bottomRow+1))
	seq.WriteString(fmt.Sprintf("\033[%d;1H", bottomRow+1))
	for i := 0; i < n; i++ {
		seq.WriteString("\n")
	}
	seq.WriteString("\033[r")
}

func inlineAppendClearRows(seq *strings.Builder, startRow, count int) {
	if startRow < 0 || count < 1 {
		return
	}
	for i := 0; i < count; i++ {
		row := startRow + i
		seq.WriteString(fmt.Sprintf("\033[%d;1H\033[2K", row+1))
	}
}
