package main

import (
	"time"

	tui "github.com/grindlemire/go-tui"
)

// Message represents a chat message
type Message struct {
	Role      string        // "user" | "assistant" | "system"
	Content   string
	Tokens    int
	Duration  time.Duration
	Timestamp time.Time
	Streaming bool // true while still receiving tokens
}

// ChatEvent for cross-component communication
type ChatEvent struct {
	Type    string // "token" | "done" | "error" | "cancel"
	Payload string
}

// AppState holds all shared application state
type AppState struct {
	// Provider configuration
	Provider     *tui.State[string]
	Model        *tui.State[string]
	Temperature  *tui.State[float64]
	SystemPrompt *tui.State[string]

	// Available options (populated on init)
	AvailableProviders []string
	ProviderModels     map[string][]string

	// Conversation
	Messages *tui.State[[]Message]

	// UI state
	TotalTokens *tui.State[int]
	IsStreaming *tui.State[bool]
	Error       *tui.State[string]
}

// NewAppState creates initialized app state with defaults
func NewAppState() *AppState {
	return &AppState{
		Provider:     tui.NewState("openai"),
		Model:        tui.NewState("gpt-4"),
		Temperature:  tui.NewState(0.7),
		SystemPrompt: tui.NewState("You are a helpful assistant."),

		AvailableProviders: []string{},
		ProviderModels: map[string][]string{
			"openai":    {"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
			"anthropic": {"claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"},
			"ollama":    {"llama2", "mistral", "codellama"},
		},

		Messages:    tui.NewState([]Message{}),
		TotalTokens: tui.NewState(0),
		IsStreaming: tui.NewState(false),
		Error:       tui.NewState(""),
	}
}

// AddMessage appends a message to the conversation
func (s *AppState) AddMessage(msg Message) {
	msgs := s.Messages.Get()
	s.Messages.Set(append(msgs, msg))
}

// UpdateLastMessage updates the last message (for streaming)
func (s *AppState) UpdateLastMessage(content string, done bool) {
	msgs := s.Messages.Get()
	if len(msgs) == 0 {
		return
	}
	msgs[len(msgs)-1].Content = content
	msgs[len(msgs)-1].Streaming = !done
	s.Messages.Set(msgs)
}

// ClearMessages resets the conversation
func (s *AppState) ClearMessages() {
	s.Messages.Set([]Message{})
	s.TotalTokens.Set(0)
}
