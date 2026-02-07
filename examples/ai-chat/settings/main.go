package settings

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

// SettingsResult contains the updated settings when saved
type SettingsResult struct {
	Provider     string
	Model        string
	Temperature  float64
	SystemPrompt string
	Saved        bool
}

// Show displays the settings screen in alternate buffer and returns results
func Show(
	currentProvider string,
	currentModel string,
	currentTemp float64,
	currentPrompt string,
	availableProviders []string,
	providerModels map[string][]string,
) SettingsResult {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create settings app: %v\n", err)
		return SettingsResult{Saved: false}
	}
	defer app.Close()

	// Create settings state
	state := &SettingsState{
		Provider:           tui.NewState(currentProvider),
		Model:              tui.NewState(currentModel),
		Temperature:        tui.NewState(currentTemp),
		SystemPrompt:       tui.NewState(currentPrompt),
		AvailableProviders: availableProviders,
		ProviderModels:     providerModels,
		FocusedSection:     tui.NewState(0),
		Saved:              tui.NewState(false),
	}

	app.SetRoot(SettingsApp(state))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Settings error: %v\n", err)
	}

	return SettingsResult{
		Provider:     state.Provider.Get(),
		Model:        state.Model.Get(),
		Temperature:  state.Temperature.Get(),
		SystemPrompt: state.SystemPrompt.Get(),
		Saved:        state.Saved.Get(),
	}
}

// SettingsState holds settings form state
type SettingsState struct {
	Provider           *tui.State[string]
	Model              *tui.State[string]
	Temperature        *tui.State[float64]
	SystemPrompt       *tui.State[string]
	AvailableProviders []string
	ProviderModels     map[string][]string
	FocusedSection     *tui.State[int] // 0=provider, 1=model, 2=temp, 3=prompt
	Saved              *tui.State[bool]
}
