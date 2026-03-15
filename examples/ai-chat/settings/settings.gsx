package settings

import (
	"fmt"
	"strings"
	tui "github.com/grindlemire/go-tui"
)

const (
	numSections        = 4
	minTurns           = 1
	maxTurns           = 50
	turnsBarWidth      = 22
	valuePreviewWidth  = 30
	promptPreviewWidth = 48
	promptPreviewLines = 2
)

type SettingsApp struct {
	Model               *tui.State[string]
	MaxTurns            *tui.State[int]
	PermissionMode      *tui.State[string]
	SystemPrompt        *tui.State[string]
	SystemPromptPresets []string
	AvailableModels     []string
	AvailablePermModes  []string
	FocusedSection      *tui.State[int]
	onClose             func()
}

func NewSettingsApp(
	model *tui.State[string],
	maxTurns *tui.State[int],
	permissionMode *tui.State[string],
	systemPrompt *tui.State[string],
	onClose func(),
) *SettingsApp {
	return &SettingsApp{
		Model:               model,
		MaxTurns:            maxTurns,
		PermissionMode:      permissionMode,
		SystemPrompt:        systemPrompt,
		SystemPromptPresets: buildSystemPromptPresets(systemPrompt.Get()),
		AvailableModels:     []string{"sonnet", "opus", "haiku"},
		AvailablePermModes:  []string{"default", "plan"},
		FocusedSection:      tui.NewState(0),
		onClose:             onClose,
	}
}

func (s *SettingsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRuneMod('s', tui.ModCtrl, func(ke tui.KeyEvent) { s.close() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { s.close() }),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { s.close() }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { s.nextSection() }),
		tui.OnKeyStop(tui.KeyLeft, func(ke tui.KeyEvent) { s.handleLeft() }),
		tui.OnKeyStop(tui.KeyRight, func(ke tui.KeyEvent) { s.handleRight() }),
		tui.OnKeyStop(tui.KeyUp, func(ke tui.KeyEvent) { s.handleUp() }),
		tui.OnKeyStop(tui.KeyDown, func(ke tui.KeyEvent) { s.handleDown() }),
		tui.OnRune('h', func(ke tui.KeyEvent) { s.handleLeft() }),
		tui.OnRune('l', func(ke tui.KeyEvent) { s.handleRight() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { s.handleUp() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { s.handleDown() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { s.close() }),
	}
}

func (s *SettingsApp) close() {
	if s.onClose != nil {
		s.onClose()
	}
}

func (s *SettingsApp) nextSection() {
	next := s.FocusedSection.Get() + 1
	if next >= numSections {
		next = 0
	}
	s.FocusedSection.Set(next)
}

func (s *SettingsApp) handleLeft() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleModel(-1)
	case 1:
		s.adjustMaxTurns(-1)
	case 2:
		s.cyclePermMode(-1)
	case 3:
		s.cycleSystemPrompt(-1)
	}
}

func (s *SettingsApp) handleRight() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleModel(1)
	case 1:
		s.adjustMaxTurns(1)
	case 2:
		s.cyclePermMode(1)
	case 3:
		s.cycleSystemPrompt(1)
	}
}

func (s *SettingsApp) handleUp() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleModel(-1)
	case 1:
		s.adjustMaxTurns(1)
	case 2:
		s.cyclePermMode(-1)
	case 3:
		s.cycleSystemPrompt(-1)
	}
}

func (s *SettingsApp) handleDown() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleModel(1)
	case 1:
		s.adjustMaxTurns(-1)
	case 2:
		s.cyclePermMode(1)
	case 3:
		s.cycleSystemPrompt(1)
	}
}

func (s *SettingsApp) cycleModel(dir int) {
	if len(s.AvailableModels) == 0 {
		return
	}

	current := s.Model.Get()
	idx := 0
	for i, m := range s.AvailableModels {
		if m == current {
			idx = i
			break
		}
	}

	idx = wrapIndex(idx+dir, len(s.AvailableModels))
	s.Model.Set(s.AvailableModels[idx])
}

func (s *SettingsApp) adjustMaxTurns(delta int) {
	t := s.MaxTurns.Get() + delta
	if t < minTurns {
		t = minTurns
	}
	if t > maxTurns {
		t = maxTurns
	}
	s.MaxTurns.Set(t)
}

func (s *SettingsApp) cyclePermMode(dir int) {
	if len(s.AvailablePermModes) == 0 {
		return
	}

	current := s.PermissionMode.Get()
	idx := 0
	for i, m := range s.AvailablePermModes {
		if m == current {
			idx = i
			break
		}
	}

	idx = wrapIndex(idx+dir, len(s.AvailablePermModes))
	s.PermissionMode.Set(s.AvailablePermModes[idx])
}

func (s *SettingsApp) cycleSystemPrompt(dir int) {
	if len(s.SystemPromptPresets) == 0 {
		return
	}

	s.ensurePromptPresetRegistered()
	current := s.SystemPrompt.Get()
	idx := indexOfExact(s.SystemPromptPresets, current)
	if idx < 0 {
		return
	}

	idx = wrapIndex(idx+dir, len(s.SystemPromptPresets))
	s.SystemPrompt.Set(s.SystemPromptPresets[idx])
}

func (s *SettingsApp) sectionAccentColor(section int) tui.Color {
	switch section {
	case 0:
		return tui.BrightCyan
	case 1:
		return tui.BrightBlue
	case 2:
		return tui.BrightYellow
	default:
		return tui.BrightGreen
	}
}

func (s *SettingsApp) sectionBorder(section int) tui.BorderStyle {
	if s.isFocused(section) {
		return tui.BorderDouble
	}
	return tui.BorderRounded
}

func (s *SettingsApp) borderStyleForSection(section int) tui.Style {
	if s.isFocused(section) {
		return tui.NewStyle().Foreground(s.sectionAccentColor(section)).Bold()
	}
	return tui.NewStyle().Foreground(tui.BrightBlack)
}

func (s *SettingsApp) sectionTitleStyle(section int) tui.Style {
	if s.isFocused(section) {
		return tui.NewStyle().Bold().Foreground(s.sectionAccentColor(section))
	}
	return tui.NewStyle().Bold().Foreground(tui.BrightWhite)
}

func (s *SettingsApp) sectionValueStyle(section int) tui.Style {
	if s.isFocused(section) {
		return tui.NewStyle().Bold().Foreground(s.sectionAccentColor(section))
	}
	return tui.NewStyle().Foreground(tui.BrightWhite)
}

func (s *SettingsApp) isFocused(section int) bool {
	return s.FocusedSection.Get() == section
}

func (s *SettingsApp) fieldLabel(section int, label string) string {
	if s.isFocused(section) {
		return "› " + label
	}
	return "  " + label
}

func (s *SettingsApp) modelOptionLabel(model string) string {
	if model == s.Model.Get() {
		return "  ● " + model
	}
	return "  ○ " + model
}

func (s *SettingsApp) modelOptionStyle(model string) tui.Style {
	if model == s.Model.Get() {
		if s.isFocused(0) {
			return tui.NewStyle().Bold().Foreground(s.sectionAccentColor(0))
		}
		return tui.NewStyle().Bold().Foreground(tui.Cyan)
	}
	return tui.NewStyle().Foreground(tui.White).Dim()
}

func (s *SettingsApp) permModeOptionLabel(mode string) string {
	if mode == s.PermissionMode.Get() {
		return "  ● " + mode
	}
	return "  ○ " + mode
}

func (s *SettingsApp) permModeOptionStyle(mode string) tui.Style {
	if mode == s.PermissionMode.Get() {
		if s.isFocused(2) {
			return tui.NewStyle().Bold().Foreground(s.sectionAccentColor(2))
		}
		return tui.NewStyle().Bold().Foreground(tui.Yellow)
	}
	return tui.NewStyle().Foreground(tui.White).Dim()
}

func (s *SettingsApp) modelSummary() string {
	if len(s.AvailableModels) == 0 {
		return "none"
	}

	index := indexOf(s.AvailableModels, s.Model.Get()) + 1
	return indexedSummary(s.Model.Get(), index, len(s.AvailableModels))
}

func (s *SettingsApp) maxTurnsSummary() string {
	return fmt.Sprintf("%d", s.MaxTurns.Get())
}

func (s *SettingsApp) permModeSummary() string {
	if len(s.AvailablePermModes) == 0 {
		return "none"
	}

	index := indexOf(s.AvailablePermModes, s.PermissionMode.Get()) + 1
	return indexedSummary(s.PermissionMode.Get(), index, len(s.AvailablePermModes))
}

func (s *SettingsApp) promptPresetSummary() string {
	if len(s.SystemPromptPresets) == 0 {
		return "preset 0/0"
	}

	index := indexOfExact(s.SystemPromptPresets, s.SystemPrompt.Get())
	if index < 0 {
		return fmt.Sprintf("preset ?/%d", len(s.SystemPromptPresets))
	}
	return fmt.Sprintf("preset %d/%d", index+1, len(s.SystemPromptPresets))
}

func (s *SettingsApp) activeSectionHint() string {
	switch s.FocusedSection.Get() {
	case 0:
		return "Model focus: ↑/↓ or ←/→ cycles models"
	case 1:
		return "Max turns focus: ↑/→ increases, ↓/← decreases"
	case 2:
		return "Permission mode focus: ↑/↓ or ←/→ cycles modes"
	default:
		return "System prompt focus: ↑/↓ cycles prompt presets"
	}
}

func (s *SettingsApp) maxTurnsBar() string {
	t := s.MaxTurns.Get()
	pos := int(float64(t-minTurns) / float64(maxTurns-minTurns) * float64(turnsBarWidth-1) + 0.5)
	if pos < 0 {
		pos = 0
	}
	if pos >= turnsBarWidth {
		pos = turnsBarWidth - 1
	}

	var bar strings.Builder
	bar.Grow(turnsBarWidth * 3)
	for i := 0; i < turnsBarWidth; i++ {
		if i == pos {
			bar.WriteString("●")
		} else {
			bar.WriteString("━")
		}
	}
	return bar.String()
}

func (s *SettingsApp) promptPreview() []string {
	normalized := strings.ReplaceAll(s.SystemPrompt.Get(), "\n", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if normalized == "" {
		return []string{"(empty)"}
	}

	lines := wrapWords(normalized, promptPreviewWidth)
	if len(lines) <= promptPreviewLines {
		return lines
	}

	clipped := append([]string{}, lines[:promptPreviewLines]...)
	clipped[promptPreviewLines-1] = truncateRunes(clipped[promptPreviewLines-1], promptPreviewWidth-1) + "…"
	return clipped
}

func wrapWords(text string, width int) []string {
	if width <= 0 {
		return nil
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, len(words))
	current := ""

	for _, word := range words {
		for runeLen(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}

			head, tail := splitAtRunes(word, width)
			lines = append(lines, head)
			word = tail
		}

		if current == "" {
			current = word
			continue
		}

		candidate := current + " " + word
		if runeLen(candidate) <= width {
			current = candidate
			continue
		}

		lines = append(lines, current)
		current = word
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func splitAtRunes(text string, width int) (string, string) {
	runes := []rune(text)
	if len(runes) <= width {
		return text, ""
	}
	return string(runes[:width]), string(runes[width:])
}

func buildSystemPromptPresets(current string) []string {
	base := []string{
		"You are a helpful assistant.",
		"You are a concise assistant. Give practical steps and short explanations.",
		"You are an expert coding assistant. Prioritize correctness, tradeoffs, and tests.",
		"You are a creative collaborator. Offer multiple options before recommending one.",
	}

	presets := make([]string, 0, len(base)+1)
	if strings.TrimSpace(current) != "" {
		presets = append(presets, current)
	}

	for _, prompt := range base {
		if indexOfExact(presets, prompt) >= 0 {
			continue
		}
		presets = append(presets, prompt)
	}

	return presets
}

func indexedSummary(value string, index, total int) string {
	suffix := fmt.Sprintf(" (%d/%d)", index, total)
	valueWidth := valuePreviewWidth - runeLen(suffix)
	if valueWidth < 1 {
		valueWidth = 1
	}
	return truncateRunes(value, valueWidth) + suffix
}

func truncateRunes(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

func runeLen(text string) int {
	return len([]rune(text))
}

func (s *SettingsApp) ensurePromptPresetRegistered() {
	current := s.SystemPrompt.Get()
	if strings.TrimSpace(current) == "" {
		return
	}
	if indexOfExact(s.SystemPromptPresets, current) >= 0 {
		return
	}
	s.SystemPromptPresets = append([]string{current}, s.SystemPromptPresets...)
}

func indexOf(items []string, value string) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return 0
}

func indexOfExact(items []string, value string) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return -1
}

func wrapIndex(idx, length int) int {
	for idx < 0 {
		idx += length
	}
	for idx >= length {
		idx -= length
	}
	return idx
}

templ (s *SettingsApp) Render() {
	<div class="flex-col h-full p-1 gap-0">
		<div class="flex-col items-center shrink-0 border-double border-gradient-cyan-blue p-1">
			<span class="text-gradient-bright-cyan-bright-yellow font-bold">{"Claude Settings"}</span>
			<span class="text-bright-cyan">{"Model, turn limits, permissions, and system prompt"}</span>
		</div>

		<div class="shrink-0 border-gradient-cyan-blue" border={s.sectionBorder(0)} borderStyle={s.borderStyleForSection(0)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(0)}>{s.fieldLabel(0, "Model")}</span>
					<span textStyle={s.sectionValueStyle(0)}>{s.modelSummary()}</span>
				</div>

				<div class="flex gap-2">
					for _, model := range s.AvailableModels {
						<span textStyle={s.modelOptionStyle(model)}>{s.modelOptionLabel(model)}</span>
					}
				</div>
			</div>
		</div>

		<div class="shrink-0 border-gradient-blue-cyan" border={s.sectionBorder(1)} borderStyle={s.borderStyleForSection(1)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(1)}>{s.fieldLabel(1, "Max Turns")}</span>
					<span textStyle={s.sectionValueStyle(1)}>{s.maxTurnsSummary()}</span>
				</div>

				<span class="text-gradient-blue-cyan">{s.maxTurnsBar()}</span>
			</div>
		</div>

		<div class="shrink-0 border-gradient-yellow-cyan" border={s.sectionBorder(2)} borderStyle={s.borderStyleForSection(2)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(2)}>{s.fieldLabel(2, "Permission Mode")}</span>
					<span textStyle={s.sectionValueStyle(2)}>{s.permModeSummary()}</span>
				</div>

				<div class="flex gap-2">
					for _, mode := range s.AvailablePermModes {
						<span textStyle={s.permModeOptionStyle(mode)}>{s.permModeOptionLabel(mode)}</span>
					}
				</div>
			</div>
		</div>

		<div class="border-gradient-green-cyan" border={s.sectionBorder(3)} borderStyle={s.borderStyleForSection(3)} flexGrow={1}>
			<div class="flex-col gap-1">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(3)}>{s.fieldLabel(3, "System Prompt")}</span>
					<span textStyle={s.sectionValueStyle(3)}>{s.promptPresetSummary()}</span>
				</div>

				for _, line := range s.promptPreview() {
					<span class="text-white">{line}</span>
				}
			</div>
		</div>

		<div class="flex-col items-center shrink-0 border-thick border-gradient-white-black">
			<span class="font-dim">{"Tab: next section   arrows or h/j/k/l: change   Enter/Esc/Ctrl+S/q: close"}</span>
			<span class="text-bright-cyan">{s.activeSectionHint()}</span>
		</div>
	</div>
}
