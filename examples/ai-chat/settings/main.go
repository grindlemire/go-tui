package settings

import (
	tui "github.com/grindlemire/go-tui"
)

// SettingsState holds the settings form state
type SettingsState struct {
	// References to main app state (modified directly)
	Provider     *tui.State[string]
	Model        *tui.State[string]
	Temperature  *tui.State[float64]
	SystemPrompt *tui.State[string]

	// Available options
	AvailableProviders []string
	ProviderModels     map[string][]string

	// Local UI state
	FocusedSection *tui.State[int]

	// Callback to close settings
	OnClose func()
}

const (
	numSections = 4
	minTemp     = 0.0
	maxTemp     = 1.0
)

// SettingsApp is the settings component
type SettingsApp struct {
	state     *SettingsState
	saveBtn   *tui.Ref
	cancelBtn *tui.Ref
}

// NewSettingsApp creates a new settings component
func NewSettingsApp(provider *tui.State[string], model *tui.State[string], temperature *tui.State[float64], systemPrompt *tui.State[string], availableProviders []string, providerModels map[string][]string, onClose func()) *SettingsApp {
	return &SettingsApp{
		state: &SettingsState{
			Provider:           provider,
			Model:              model,
			Temperature:        temperature,
			SystemPrompt:       systemPrompt,
			AvailableProviders: availableProviders,
			ProviderModels:     providerModels,
			FocusedSection:     tui.NewState(0),
			OnClose:            onClose,
		},
		saveBtn:   tui.NewRef(),
		cancelBtn: tui.NewRef(),
	}
}

func (s *SettingsApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
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

func (s *SettingsApp) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(s.saveBtn, s.close),
		tui.Click(s.cancelBtn, s.close),
	)
}

func (s *SettingsApp) close() {
	if s.state.OnClose != nil {
		s.state.OnClose()
	}
}

func (s *SettingsApp) nextSection() {
	next := s.state.FocusedSection.Get() + 1
	if next >= numSections {
		next = 0
	}
	s.state.FocusedSection.Set(next)
}

func (s *SettingsApp) handleLeft() {
	section := s.state.FocusedSection.Get()
	switch section {
	case 0:
		s.cycleProvider(-1)
	case 1:
		s.cycleModel(-1)
	case 2:
		s.adjustTemp(-0.1)
	}
}

func (s *SettingsApp) handleRight() {
	section := s.state.FocusedSection.Get()
	switch section {
	case 0:
		s.cycleProvider(1)
	case 1:
		s.cycleModel(1)
	case 2:
		s.adjustTemp(0.1)
	}
}

func (s *SettingsApp) cycleProvider(dir int) {
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

func (s *SettingsApp) cycleModel(dir int) {
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

func (s *SettingsApp) adjustTemp(delta float64) {
	t := s.state.Temperature.Get() + delta
	if t < minTemp {
		t = minTemp
	} else if t > maxTemp {
		t = maxTemp
	}
	s.state.Temperature.Set(t)
}

func (s *SettingsApp) borderStyleForSection(section int) tui.Style {
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

func (s *SettingsApp) tempBar() string {
	t := s.state.Temperature.Get()
	pos := int(t * 29)
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

// Run runs the settings as a standalone fullscreen app.
// This blocks until the user closes settings.
func Run(provider *tui.State[string], model *tui.State[string], temperature *tui.State[float64], systemPrompt *tui.State[string], availableProviders []string, providerModels map[string][]string) error {
	app, err := tui.NewApp() // Fullscreen mode (no WithInlineHeight)
	if err != nil {
		return err
	}
	defer app.Close()

	settings := NewSettingsApp(provider, model, temperature, systemPrompt, availableProviders, providerModels, func() {
		tui.Stop() // Stop the settings app when closed
	})

	app.SetRoot(settings)
	return app.Run()
}
