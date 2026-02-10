package settings

import (
	"fmt"
	"strings"
	tui "github.com/grindlemire/go-tui"
)

const (
	numSections        = 4
	minTemp            = 0.0
	maxTemp            = 1.0
	tempBarWidth       = 22
	valuePreviewWidth  = 30
	promptPreviewWidth = 48
	promptPreviewLines = 2
)

type SettingsApp struct {
	Provider           *tui.State[string]
	Model              *tui.State[string]
	Temperature        *tui.State[float64]
	SystemPrompt       *tui.State[string]
	SystemPromptPresets []string
	AvailableProviders []string
	ProviderModels     map[string][]string
	FocusedSection     *tui.State[int]
	onClose            func()
}

func NewSettingsApp(
	provider *tui.State[string], 
	model *tui.State[string], 
	temperature *tui.State[float64], 
	systemPrompt *tui.State[string], 
	availableProviders []string, 
	providerModels map[string][]string, 
	onClose func(),
	) *SettingsApp {
	return &SettingsApp{
		Provider:            provider,
		Model:               model,
		Temperature:         temperature,
		SystemPrompt:        systemPrompt,
		SystemPromptPresets: buildSystemPromptPresets(systemPrompt.Get()),
		AvailableProviders:  availableProviders,
		ProviderModels:      providerModels,
		FocusedSection:      tui.NewState(0),
		onClose:             onClose,
	}
}

func (s *SettingsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyCtrlS, func(ke tui.KeyEvent) { s.close() }),
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
		s.cycleProvider(-1)
	case 1:
		s.cycleModel(-1)
	case 2:
		s.adjustTemp(-0.1)
	case 3:
		s.cycleSystemPrompt(-1)
	}
}

func (s *SettingsApp) handleRight() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleProvider(1)
	case 1:
		s.cycleModel(1)
	case 2:
		s.adjustTemp(0.1)
	case 3:
		s.cycleSystemPrompt(1)
	}
}

func (s *SettingsApp) handleUp() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleProvider(-1)
	case 1:
		s.cycleModel(-1)
	case 2:
		s.adjustTemp(0.1)
	case 3:
		s.cycleSystemPrompt(-1)
	}
}

func (s *SettingsApp) handleDown() {
	switch s.FocusedSection.Get() {
	case 0:
		s.cycleProvider(1)
	case 1:
		s.cycleModel(1)
	case 2:
		s.adjustTemp(-0.1)
	case 3:
		s.cycleSystemPrompt(1)
	}
}

func (s *SettingsApp) cycleProvider(dir int) {
	if len(s.AvailableProviders) == 0 {
		return
	}

	current := s.Provider.Get()
	idx := 0
	for i, p := range s.AvailableProviders {
		if p == current {
			idx = i
			break
		}
	}

	idx = wrapIndex(idx+dir, len(s.AvailableProviders))
	nextProvider := s.AvailableProviders[idx]
	s.Provider.Set(nextProvider)

	models := s.ProviderModels[nextProvider]
	if len(models) > 0 {
		s.Model.Set(models[0])
	}
}

func (s *SettingsApp) cycleModel(dir int) {
	models := s.ProviderModels[s.Provider.Get()]
	if len(models) == 0 {
		return
	}

	current := s.Model.Get()
	idx := 0
	for i, m := range models {
		if m == current {
			idx = i
			break
		}
	}

	idx = wrapIndex(idx+dir, len(models))
	s.Model.Set(models[idx])
}

func (s *SettingsApp) adjustTemp(delta float64) {
	t := s.Temperature.Get() + delta
	if t < minTemp {
		t = minTemp
	}
	if t > maxTemp {
		t = maxTemp
	}
	s.Temperature.Set(t)
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

func (s *SettingsApp) providerOptionLabel(provider string) string {
	if provider == s.Provider.Get() {
		return "  ● " + provider
	}
	return "  ○ " + provider
}

func (s *SettingsApp) providerOptionStyle(provider string) tui.Style {
	if provider == s.Provider.Get() {
		if s.isFocused(0) {
			return tui.NewStyle().Bold().Foreground(s.sectionAccentColor(0))
		}
		return tui.NewStyle().Bold().Foreground(tui.Cyan)
	}
	return tui.NewStyle().Foreground(tui.White).Dim()
}

func (s *SettingsApp) providerSummary() string {
	if len(s.AvailableProviders) == 0 {
		return "none"
	}

	index := indexOf(s.AvailableProviders, s.Provider.Get()) + 1
	return indexedSummary(s.Provider.Get(), index, len(s.AvailableProviders))
}

func (s *SettingsApp) modelSummary() string {
	models := s.ProviderModels[s.Provider.Get()]
	if len(models) == 0 {
		return "none"
	}

	index := indexOf(models, s.Model.Get()) + 1
	return indexedSummary(s.Model.Get(), index, len(models))
}

func (s *SettingsApp) temperatureSummary() string {
	return fmt.Sprintf("%.1f  %s", s.Temperature.Get(), s.temperatureMode())
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

func (s *SettingsApp) temperatureMode() string {
	t := s.Temperature.Get()
	switch {
	case t <= 0.3:
		return "precise"
	case t < 0.8:
		return "balanced"
	default:
		return "creative"
	}
}

func (s *SettingsApp) activeSectionHint() string {
	switch s.FocusedSection.Get() {
	case 0:
		return "Provider focus: ↑/↓ cycles providers"
	case 1:
		return "Model focus: ↑/↓ or ←/→ cycles models"
	case 2:
		return "Temperature focus: ↑ increases, ↓ decreases"
	default:
		return "System prompt focus: ↑/↓ cycles prompt presets"
	}
}

func (s *SettingsApp) tempBar() string {
	t := s.Temperature.Get()
	pos := int(t*float64(tempBarWidth-1) + 0.5)
	if pos < 0 {
		pos = 0
	}
	if pos >= tempBarWidth {
		pos = tempBarWidth - 1
	}

	var bar strings.Builder
	bar.Grow(tempBarWidth * 3)
	for i := 0; i < tempBarWidth; i++ {
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
			<span class="text-gradient-bright-cyan-bright-yellow font-bold">{"AI Chat Settings"}</span>
			<span class="text-bright-cyan">{"Control center for provider, model, temperature, and prompt"}</span>
		</div>

		<div class="shrink-0 border-gradient-cyan-blue" border={s.sectionBorder(0)} borderStyle={s.borderStyleForSection(0)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(0)}>{s.fieldLabel(0, "Provider")}</span>
					<span textStyle={s.sectionValueStyle(0)}>{s.providerSummary()}</span>
				</div>

				<div class="flex gap-2">
					@for _, provider := range s.AvailableProviders {
						<span textStyle={s.providerOptionStyle(provider)}>{s.providerOptionLabel(provider)}</span>
					}
				</div>
			</div>
		</div>

		<div class="shrink-0 border-gradient-blue-cyan" border={s.sectionBorder(1)} borderStyle={s.borderStyleForSection(1)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(1)}>{s.fieldLabel(1, "Model")}</span>
					<span textStyle={s.sectionValueStyle(1)}>{s.modelSummary()}</span>
				</div>
			</div>
		</div>

		<div class="shrink-0 border-gradient-yellow-cyan" border={s.sectionBorder(2)} borderStyle={s.borderStyleForSection(2)}>
			<div class="flex-col">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(2)}>{s.fieldLabel(2, "Temperature")}</span>
					<span textStyle={s.sectionValueStyle(2)}>{s.temperatureSummary()}</span>
				</div>

				<span class="text-gradient-yellow-cyan">{s.tempBar()}</span>
			</div>
		</div>

		<div class="border-gradient-green-cyan" border={s.sectionBorder(3)} borderStyle={s.borderStyleForSection(3)} flexGrow={1}>
			<div class="flex-col gap-1">
				<div class="flex justify-between items-center">
					<span textStyle={s.sectionTitleStyle(3)}>{s.fieldLabel(3, "System Prompt")}</span>
					<span textStyle={s.sectionValueStyle(3)}>{s.promptPresetSummary()}</span>
				</div>

				@for _, line := range s.promptPreview() {
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
