package main

import (
	"fmt"
	"os"
	"sort"

	tui "github.com/grindlemire/go-tui"
)

// Generate with: tui generate . (from parent dir: go run ./cmd/tui generate ./examples/ai-chat/...)

func main() {
	providers := DetectProviders()
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No providers available. Set one of:")
		fmt.Fprintln(os.Stderr, "  OPENAI_API_KEY")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_API_KEY")
		fmt.Fprintln(os.Stderr, "  Or have Ollama running locally")
		os.Exit(1)
	}

	state := NewAppState()

	// Collect and sort provider names for consistent ordering
	for name := range providers {
		state.AvailableProviders = append(state.AvailableProviders, name)
	}
	sort.Strings(state.AvailableProviders)

	// Set default provider and model
	if len(state.AvailableProviders) > 0 {
		state.Provider.Set(state.AvailableProviders[0])
		models := state.ProviderModels[state.AvailableProviders[0]]
		if len(models) > 0 {
			state.Model.Set(models[0])
		}
	}

	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetRoot(ChatApp(state, providers))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
