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

type settingsApp struct {
	state     *SettingsState
	saveBtn   *tui.Ref
	cancelBtn *tui.Ref
}

func SettingsApp(state *SettingsState) *settingsApp {
	return &settingsApp{
		state:     state,
		saveBtn:   tui.NewRef(),
		cancelBtn: tui.NewRef(),
	}
}

func (s *settingsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { tui.Stop() }),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { s.save() }),
		tui.OnKey(tui.KeyTab, func(ke tui.KeyEvent) { s.nextSection() }),
		tui.OnKeyStop(tui.KeyLeft, func(ke tui.KeyEvent) { s.handleLeft() }),
		tui.OnKeyStop(tui.KeyRight, func(ke tui.KeyEvent) { s.handleRight() }),
		tui.OnRune('h', func(ke tui.KeyEvent) { s.handleLeft() }),
		tui.OnRune('l', func(ke tui.KeyEvent) { s.handleRight() }),
	}
}

func (s *settingsApp) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(s.saveBtn, s.save),
		tui.Click(s.cancelBtn, func() { tui.Stop() }),
	)
}

func (s *settingsApp) save() {
	s.state.Saved.Set(true)
	tui.Stop()
}

func (s *settingsApp) nextSection() {
	next := s.state.FocusedSection.Get() + 1
	if next >= numSections {
		next = 0
	}
	s.state.FocusedSection.Set(next)
}

func (s *settingsApp) handleLeft() {
	section := s.state.FocusedSection.Get()
	switch section {
	case 0: // Provider
		s.cycleProvider(-1)
	case 1: // Model
		s.cycleModel(-1)
	case 2: // Temperature
		s.adjustTemp(-0.1)
	case 3:
		// System Prompt section: no action on left/right
	}
}

func (s *settingsApp) handleRight() {
	section := s.state.FocusedSection.Get()
	switch section {
	case 0:
		s.cycleProvider(1)
	case 1:
		s.cycleModel(1)
	case 2:
		s.adjustTemp(0.1)
	case 3:
		// System Prompt section: no action on left/right
	}
}

func (s *settingsApp) cycleProvider(dir int) {
	providers := s.state.AvailableProviders
	if len(providers) == 0 {
		return
	}
	current := s.state.Provider.Get()
	idx := 0
	for i, p := range providers {
		if p == current {
			idx = i
			break
		}
	}
	idx = idx + dir + len(providers)
	idx = wrapIndex(idx, len(providers))
	s.state.Provider.Set(providers[idx])
	// Update model to first of new provider
	models := s.state.ProviderModels[providers[idx]]
	if len(models) > 0 {
		s.state.Model.Set(models[0])
	}
}

func (s *settingsApp) cycleModel(dir int) {
	provider := s.state.Provider.Get()
	models := s.state.ProviderModels[provider]
	if len(models) == 0 {
		return
	}
	current := s.state.Model.Get()
	idx := 0
	for i, m := range models {
		if m == current {
			idx = i
			break
		}
	}
	idx = idx + dir + len(models)
	idx = wrapIndex(idx, len(models))
	s.state.Model.Set(models[idx])
}

func (s *settingsApp) adjustTemp(delta float64) {
	t := s.state.Temperature.Get() + delta
	if t < minTemp {
		t = minTemp
	} else if t > maxTemp {
		t = maxTemp
	}
	s.state.Temperature.Set(t)
}

// Returns the border style for a section based on focus
func (s *settingsApp) borderStyleForSection(section int) tui.Style {
	if s.state.FocusedSection.Get() == section {
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

func (s *settingsApp) tempBar() string {
	t := s.state.Temperature.Get()
	pos := int(t * 29) // position 0-29 for 30 character bar
	bar := ""
	for i := 0; i < 30; i++ {
		if i < pos {
			bar += "━"
		} else if i == pos {
			bar += "●"
		} else {
			bar += "━"
		}
	}
	return bar
}

templ (s *settingsApp) Render() {
	<div class="flex-col h-full p-2 gap-2">
		<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifyCenter} align={tui.AlignCenter}>
			<span class="text-gradient-cyan-magenta font-bold">{"  Settings"}</span>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(0)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Provider"}</span>
				<div class="flex gap-2">
					@for _, p := range s.state.AvailableProviders {
						@if p == s.state.Provider.Get() {
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
					@for _, m := range s.state.ProviderModels[s.state.Provider.Get()] {
						@if m == s.state.Model.Get() {
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
					<span class="text-cyan">{fmt.Sprintf("%.1f", s.state.Temperature.Get())}</span>
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
				<span class="text-white">{s.state.SystemPrompt.Get()}</span>
			</div>
		</div>

		<div class="flex justify-center gap-2">
			<button ref={s.saveBtn} class="border-rounded border-cyan p-1">{"  Save  "}</button>
			<button ref={s.cancelBtn} class="border-rounded p-1">{"  Cancel  "}</button>
		</div>

		<div class="flex justify-center">
			<span class="font-dim">{"Tab: navigate  ←/→: select  Enter: save  Esc: cancel"}</span>
		</div>
	</div>
}
