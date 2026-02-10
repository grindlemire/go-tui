package settings

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

const (
	numSections = 4
	minTemp     = 0.0
	maxTemp     = 1.0
)

type SettingsApp struct {
	Provider           *tui.State[string]
	Model              *tui.State[string]
	Temperature        *tui.State[float64]
	SystemPrompt       *tui.State[string]
	AvailableProviders []string
	ProviderModels     map[string][]string
	FocusedSection     *tui.State[int]
	onClose            func()
}

func NewSettingsApp(provider *tui.State[string], model *tui.State[string], temperature *tui.State[float64], systemPrompt *tui.State[string], availableProviders []string, providerModels map[string][]string, onClose func()) *SettingsApp {
	return &SettingsApp{
		Provider:           provider,
		Model:              model,
		Temperature:        temperature,
		SystemPrompt:       systemPrompt,
		AvailableProviders: availableProviders,
		ProviderModels:     providerModels,
		FocusedSection:     tui.NewState(0),
		onClose:            onClose,
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
		tui.OnRune('h', func(ke tui.KeyEvent) { s.handleLeft() }),
		tui.OnRune('l', func(ke tui.KeyEvent) { s.handleRight() }),
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

func (s *SettingsApp) borderStyleForSection(section int) tui.Style {
	if s.FocusedSection.Get() == section {
		return tui.NewStyle().Foreground(tui.Cyan)
	}
	return tui.NewStyle()
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

func (s *SettingsApp) tempBar() string {
	t := s.Temperature.Get()
	pos := int(t * 29)
	bar := ""
	for i := 0; i < 30; i++ {
		if i == pos {
			bar += "●"
		} else {
			bar += "━"
		}
	}
	return bar
}

templ (s *SettingsApp) Render() {
	<div class="flex-col h-full p-2 gap-2">
		<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifyCenter} align={tui.AlignCenter}>
			<span class="text-gradient-cyan-magenta font-bold">{"  Settings"}</span>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(0)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Provider"}</span>
				<div class="flex gap-2">
					@for _, p := range s.AvailableProviders {
						@if p == s.Provider.Get() {
							<span class="text-cyan font-bold">{"● " + p}</span>
						} @else {
							<span class="font-dim">{"○ " + p}</span>
						}
					}
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(1)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Model"}</span>
				<div class="flex gap-2">
					@for _, m := range s.ProviderModels[s.Provider.Get()] {
						@if m == s.Model.Get() {
							<span class="text-cyan font-bold">{"● " + m}</span>
						} @else {
							<span class="font-dim">{"○ " + m}</span>
						}
					}
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(2)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Temperature"}</span>
				<div class="flex gap-2 items-center">
					<span class="text-white">{s.tempBar()}</span>
					<span class="text-cyan">{fmt.Sprintf("%.1f", s.Temperature.Get())}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">{"← precise"}</span>
					<span class="font-dim">{"creative →"}</span>
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(3)} padding={1} flexGrow={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"System Prompt"}</span>
				<span class="text-white">{s.SystemPrompt.Get()}</span>
			</div>
		</div>

		<div class="flex justify-center">
			<span class="font-dim">{"Tab: navigate  ←/→: select  Ctrl+S/Esc/Enter: close"}</span>
		</div>
	</div>
}
