# AI Chat Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a polished AI chat application demonstrating go-tui's full capabilities with langchaingo streaming integration.

**Architecture:** Component-based UI with shared state via `*tui.State[T]` pointers, cross-component communication via `tui.Events[T]`, and streaming via channels + Watchers. Settings screen uses alternate buffer via `app.EnterAlternateScreen()`.

**Tech Stack:** go-tui (declarative UI), langchaingo (LLM abstraction), OpenAI/Anthropic/Ollama providers

---

## Task 1: Project Setup

**Files:**
- Create: `examples/ai-chat/go.mod`
- Create: `examples/ai-chat/main.go`

**Step 1: Create go.mod**

```bash
mkdir -p examples/ai-chat
```

Create `examples/ai-chat/go.mod`:
```go
module github.com/grindlemire/go-tui/examples/ai-chat

go 1.25.1

require (
	github.com/grindlemire/go-tui v0.0.0
	github.com/tmc/langchaingo v0.1.13
)

replace github.com/grindlemire/go-tui => ../..
```

**Step 2: Create minimal main.go**

Create `examples/ai-chat/main.go`:
```go
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Placeholder root
	root := tui.New(
		tui.WithText("AI Chat - Press q to quit"),
	)
	app.SetRoot(root)

	// Simple exit on 'q'
	app.SetKeyHandler(func(ke tui.KeyEvent) bool {
		if ke.Rune == 'q' {
			tui.Stop()
			return true
		}
		return false
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Verify it compiles and runs**

```bash
cd examples/ai-chat && go mod tidy && go build && ./ai-chat
```

Expected: Shows "AI Chat - Press q to quit", exits on 'q'

**Step 4: Commit**

```bash
gcommit -m "feat(examples): scaffold ai-chat example"
```

---

## Task 2: State and Types

**Files:**
- Create: `examples/ai-chat/state.go`

**Step 1: Create state.go with types**

Create `examples/ai-chat/state.go`:
```go
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
```

**Step 2: Verify it compiles**

```bash
cd examples/ai-chat && go build
```

Expected: Compiles without errors

**Step 3: Commit**

```bash
gcommit -m "feat(ai-chat): add state types and AppState"
```

---

## Task 3: Provider Abstraction

**Files:**
- Create: `examples/ai-chat/providers.go`

**Step 1: Create provider interface and implementations**

Create `examples/ai-chat/providers.go`:
```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// Provider interface for LLM backends
type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []Message, opts ChatOpts, tokenCh chan<- string) error
}

// ChatOpts configures a chat request
type ChatOpts struct {
	Model        string
	Temperature  float64
	SystemPrompt string
}

// --- OpenAI Provider ---

type OpenAIProvider struct {
	client llms.Model
}

func NewOpenAIProvider() (*OpenAIProvider, error) {
	client, err := openai.New()
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	return &OpenAIProvider{client: client}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, opts ChatOpts, tokenCh chan<- string) error {
	defer close(tokenCh)

	// Convert messages to langchain format
	lcMessages := make([]llms.MessageContent, 0, len(messages)+1)

	// Add system prompt
	if opts.SystemPrompt != "" {
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: opts.SystemPrompt}},
		})
	}

	for _, msg := range messages {
		role := llms.ChatMessageTypeHuman
		if msg.Role == "assistant" {
			role = llms.ChatMessageTypeAI
		}
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	_, err := p.client.GenerateContent(ctx, lcMessages,
		llms.WithModel(opts.Model),
		llms.WithTemperature(opts.Temperature),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case tokenCh <- string(chunk):
				return nil
			}
		}),
	)
	return err
}

// --- Anthropic Provider ---

type AnthropicProvider struct {
	client llms.Model
}

func NewAnthropicProvider() (*AnthropicProvider, error) {
	client, err := anthropic.New()
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}
	return &AnthropicProvider{client: client}, nil
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message, opts ChatOpts, tokenCh chan<- string) error {
	defer close(tokenCh)

	lcMessages := make([]llms.MessageContent, 0, len(messages)+1)

	if opts.SystemPrompt != "" {
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: opts.SystemPrompt}},
		})
	}

	for _, msg := range messages {
		role := llms.ChatMessageTypeHuman
		if msg.Role == "assistant" {
			role = llms.ChatMessageTypeAI
		}
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	_, err := p.client.GenerateContent(ctx, lcMessages,
		llms.WithModel(opts.Model),
		llms.WithTemperature(opts.Temperature),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case tokenCh <- string(chunk):
				return nil
			}
		}),
	)
	return err
}

// --- Ollama Provider ---

type OllamaProvider struct {
	client llms.Model
}

func NewOllamaProvider() (*OllamaProvider, error) {
	client, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	return &OllamaProvider{client: client}, nil
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message, opts ChatOpts, tokenCh chan<- string) error {
	defer close(tokenCh)

	lcMessages := make([]llms.MessageContent, 0, len(messages)+1)

	if opts.SystemPrompt != "" {
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: opts.SystemPrompt}},
		})
	}

	for _, msg := range messages {
		role := llms.ChatMessageTypeHuman
		if msg.Role == "assistant" {
			role = llms.ChatMessageTypeAI
		}
		lcMessages = append(lcMessages, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	_, err := p.client.GenerateContent(ctx, lcMessages,
		llms.WithModel(opts.Model),
		llms.WithTemperature(opts.Temperature),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case tokenCh <- string(chunk):
				return nil
			}
		}),
	)
	return err
}

// --- Provider Registry ---

// DetectProviders returns available providers based on env vars
func DetectProviders() map[string]Provider {
	providers := make(map[string]Provider)

	if os.Getenv("OPENAI_API_KEY") != "" {
		if p, err := NewOpenAIProvider(); err == nil {
			providers["openai"] = p
		}
	}

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		if p, err := NewAnthropicProvider(); err == nil {
			providers["anthropic"] = p
		}
	}

	// Ollama doesn't require API key, try to connect
	if p, err := NewOllamaProvider(); err == nil {
		providers["ollama"] = p
	}

	return providers
}
```

**Step 2: Update go.mod and verify**

```bash
cd examples/ai-chat && go mod tidy && go build
```

Expected: Compiles (may download langchaingo dependencies)

**Step 3: Commit**

```bash
gcommit -m "feat(ai-chat): add provider abstraction for OpenAI/Anthropic/Ollama"
```

---

## Task 4: Header Component

**Files:**
- Create: `examples/ai-chat/header.gsx`

**Step 1: Create header.gsx**

Create `examples/ai-chat/header.gsx`:
```gsx
package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ Header(state *AppState) {
	<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifySpaceBetween} align={tui.AlignCenter}>
		<span class="text-gradient-cyan-magenta font-bold">{"  AI Chat"}</span>
		<div class="flex gap-2">
			<span class="font-dim">{state.Model.Get()}</span>
			<span class="text-cyan">{fmt.Sprintf("%d tokens", state.TotalTokens.Get())}</span>
			<span class="font-dim">{"Ctrl+? help"}</span>
		</div>
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate header.gsx
```

Expected: Creates `header_gsx.go`

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add Header component"
```

---

## Task 5: Message Component

**Files:**
- Create: `examples/ai-chat/message.gsx`

**Step 1: Create message.gsx**

Create `examples/ai-chat/message.gsx`:
```gsx
package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type messageView struct {
	msg      Message
	index    int
	focused  bool
	events   *tui.Events[ChatEvent]
	copyBtn  *tui.Ref
	retryBtn *tui.Ref
}

func MessageView(msg Message, index int, focused bool, events *tui.Events[ChatEvent]) *messageView {
	return &messageView{
		msg:      msg,
		index:    index,
		focused:  focused,
		events:   events,
		copyBtn:  tui.NewRef(),
		retryBtn: tui.NewRef(),
	}
}

func (m *messageView) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(m.copyBtn, func() {
			m.events.Emit(ChatEvent{Type: "copy", Payload: m.msg.Content})
		}),
		tui.Click(m.retryBtn, func() {
			m.events.Emit(ChatEvent{Type: "retry", Payload: fmt.Sprintf("%d", m.index)})
		}),
	)
}

func (m *messageView) borderClass() string {
	if m.msg.Role == "assistant" {
		if m.focused {
			return "border-rounded border-cyan"
		}
		return "border-rounded border-blue"
	}
	if m.focused {
		return "border-rounded border-white"
	}
	return "border-rounded"
}

func (m *messageView) roleIcon() string {
	if m.msg.Role == "assistant" {
		return ""
	}
	return ""
}

func (m *messageView) roleClass() string {
	if m.msg.Role == "assistant" {
		return "text-cyan font-bold"
	}
	return "text-white font-bold"
}

templ (m *messageView) Render() {
	<div class={m.borderClass()} padding={1} margin={1}>
		<div class="flex-col gap-1">
			<div class="flex justify-between">
				<span class={m.roleClass()}>{m.roleIcon() + " " + m.msg.Role}</span>
				<div class="flex gap-1">
					@if m.msg.Duration > 0 {
						<span class="font-dim">{fmt.Sprintf("%.1fs", m.msg.Duration.Seconds())}</span>
					}
					@if m.msg.Role == "assistant" && !m.msg.Streaming {
						<button ref={m.retryBtn} class="font-dim">{"[r]"}</button>
					}
					<button ref={m.copyBtn} class="font-dim">{"[c]"}</button>
				</div>
			</div>
			<div>
				<span class="text-white">{m.msg.Content}</span>
				@if m.msg.Streaming {
					<span class="text-cyan font-bold">{""}</span>
				}
			</div>
		</div>
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate message.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add Message component with copy/retry actions"
```

---

## Task 6: MessageList Component

**Files:**
- Create: `examples/ai-chat/message_list.gsx`

**Step 1: Create message_list.gsx**

Create `examples/ai-chat/message_list.gsx`:
```gsx
package main

import tui "github.com/grindlemire/go-tui"

type messageList struct {
	state      *AppState
	events     *tui.Events[ChatEvent]
	focusedIdx *tui.State[int]
	content    *tui.Ref
}

func MessageList(state *AppState, events *tui.Events[ChatEvent]) *messageList {
	m := &messageList{
		state:      state,
		events:     events,
		focusedIdx: tui.NewState(-1),
		content:    tui.NewRef(),
	}

	// Subscribe to done events to scroll to bottom
	events.Subscribe(func(e ChatEvent) {
		if e.Type == "token" || e.Type == "done" {
			if el := m.content.El(); el != nil {
				el.ScrollToBottom()
			}
		}
	})

	return m
}

func (m *messageList) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('j', func(ke tui.KeyEvent) { m.moveDown() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { m.moveUp() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { m.moveDown() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { m.moveUp() }),
		tui.OnRune('g', func(ke tui.KeyEvent) { m.focusedIdx.Set(0) }),
		tui.OnRune('G', func(ke tui.KeyEvent) {
			msgs := m.state.Messages.Get()
			if len(msgs) > 0 {
				m.focusedIdx.Set(len(msgs) - 1)
			}
		}),
		tui.OnRune('c', func(ke tui.KeyEvent) { m.copyFocused() }),
		tui.OnRune('r', func(ke tui.KeyEvent) { m.retryFocused() }),
	}
}

func (m *messageList) moveDown() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx < len(msgs)-1 {
		m.focusedIdx.Set(idx + 1)
	}
}

func (m *messageList) moveUp() {
	idx := m.focusedIdx.Get()
	if idx > 0 {
		m.focusedIdx.Set(idx - 1)
	}
}

func (m *messageList) copyFocused() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx >= 0 && idx < len(msgs) {
		m.events.Emit(ChatEvent{Type: "copy", Payload: msgs[idx].Content})
	}
}

func (m *messageList) retryFocused() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx >= 0 && idx < len(msgs) && msgs[idx].Role == "assistant" {
		m.events.Emit(ChatEvent{Type: "retry"})
	}
}

templ (m *messageList) Render() {
	<div
		ref={m.content}
		class="flex-col flex-grow-1"
		scrollable={tui.ScrollVertical}
		focusable={true}>
		@for i, msg := range m.state.Messages.Get() {
			@MessageView(msg, i, i == m.focusedIdx.Get(), m.events)
		}
		@if len(m.state.Messages.Get()) == 0 {
			<div class="flex-col flex-grow-1 justify-center items-center">
				<span class="font-dim">{"No messages yet. Start typing below!"}</span>
			</div>
		}
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate message_list.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add MessageList component with vim navigation"
```

---

## Task 7: InputBar Component

**Files:**
- Create: `examples/ai-chat/input_bar.gsx`

**Step 1: Create input_bar.gsx**

Create `examples/ai-chat/input_bar.gsx`:
```gsx
package main

import tui "github.com/grindlemire/go-tui"

type inputBar struct {
	state    *AppState
	events   *tui.Events[ChatEvent]
	text     *tui.State[string]
	active   *tui.State[bool]
}

func InputBar(state *AppState, events *tui.Events[ChatEvent]) *inputBar {
	return &inputBar{
		state:  state,
		events: events,
		text:   tui.NewState(""),
		active: tui.NewState(true),
	}
}

func (i *inputBar) KeyMap() tui.KeyMap {
	if !i.active.Get() || i.state.IsStreaming.Get() {
		return nil
	}
	return tui.KeyMap{
		tui.OnRunesStop(i.appendChar),
		tui.OnKeyStop(tui.KeyBackspace, i.deleteChar),
		tui.OnKeyStop(tui.KeyEnter, i.submit),
	}
}

func (i *inputBar) appendChar(ke tui.KeyEvent) {
	i.text.Set(i.text.Get() + string(ke.Rune))
}

func (i *inputBar) deleteChar(ke tui.KeyEvent) {
	t := i.text.Get()
	if len(t) > 0 {
		// Handle UTF-8 properly
		runes := []rune(t)
		i.text.Set(string(runes[:len(runes)-1]))
	}
}

func (i *inputBar) submit(ke tui.KeyEvent) {
	t := i.text.Get()
	if t == "" {
		return
	}
	i.events.Emit(ChatEvent{Type: "submit", Payload: t})
	i.text.Set("")
}

func (i *inputBar) borderClass() string {
	if i.state.IsStreaming.Get() {
		return "border-rounded border-yellow"
	}
	return "border-rounded border-gradient-cyan-magenta"
}

templ (i *inputBar) Render() {
	<div class={i.borderClass()} height={3} padding={1} direction={tui.Row} align={tui.AlignCenter}>
		@if i.state.IsStreaming.Get() {
			<span class="text-yellow font-dim">{"  Generating..."}</span>
		} @else {
			<span class="text-cyan">{"  "}</span>
			<span class="text-white">{i.text.Get()}</span>
			<span class="text-cyan font-bold">{""}</span>
			<div class="flex-grow-1"></div>
			<span class="font-dim">{"  Enter to send"}</span>
		}
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate input_bar.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add InputBar component with text input"
```

---

## Task 8: HelpOverlay Component

**Files:**
- Create: `examples/ai-chat/help_overlay.gsx`

**Step 1: Create help_overlay.gsx**

Create `examples/ai-chat/help_overlay.gsx`:
```gsx
package main

import tui "github.com/grindlemire/go-tui"

templ HelpOverlay() {
	<div class="flex justify-center items-center" flexGrow={1}>
		<div class="border-rounded border-cyan p-2" width={50}>
			<div class="flex-col gap-1">
				<span class="text-gradient-cyan-magenta font-bold text-center">{"Keyboard Shortcuts"}</span>
				<hr />
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+,"}</span>
					<span>{"Open settings"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+?"}</span>
					<span>{"Toggle this help"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+L"}</span>
					<span>{"Clear conversation"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"Ctrl+C"}</span>
					<span>{"Cancel/Quit"}</span>
				</div>
				<hr />
				<span class="font-dim text-center">{"Message Navigation"}</span>
				<div class="flex justify-between">
					<span class="font-bold">{"j/k"}</span>
					<span>{"Move down/up"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"g/G"}</span>
					<span>{"First/Last message"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"c"}</span>
					<span>{"Copy message"}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-bold">{"r"}</span>
					<span>{"Retry response"}</span>
				</div>
				<hr />
				<span class="font-dim text-center">{"Press any key to close"}</span>
			</div>
		</div>
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate help_overlay.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add HelpOverlay component"
```

---

## Task 9: ChatApp Root Component

**Files:**
- Create: `examples/ai-chat/app.gsx`

**Step 1: Create app.gsx**

Create `examples/ai-chat/app.gsx`:
```gsx
package main

import (
	"context"
	"time"
	tui "github.com/grindlemire/go-tui"
)

type chatApp struct {
	state       *AppState
	events      *tui.Events[ChatEvent]
	providers   map[string]Provider
	showHelp    *tui.State[bool]
	tokenCh     chan string
	cancelFn    context.CancelFunc
}

func ChatApp(state *AppState, providers map[string]Provider) *chatApp {
	c := &chatApp{
		state:     state,
		events:    tui.NewEvents[ChatEvent](),
		providers: providers,
		showHelp:  tui.NewState(false),
		tokenCh:   make(chan string, 100),
	}

	// Subscribe to events
	c.events.Subscribe(c.handleEvent)

	return c
}

func (c *chatApp) handleEvent(e ChatEvent) {
	switch e.Type {
	case "submit":
		c.sendMessage(e.Payload)
	case "cancel":
		if c.cancelFn != nil {
			c.cancelFn()
		}
	case "retry":
		c.retryLast()
	}
}

func (c *chatApp) sendMessage(content string) {
	if c.state.IsStreaming.Get() {
		return
	}

	// Add user message
	c.state.AddMessage(Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})

	// Start streaming response
	c.startStreaming()
}

func (c *chatApp) startStreaming() {
	provider, ok := c.providers[c.state.Provider.Get()]
	if !ok {
		c.state.Error.Set("Provider not available")
		return
	}

	c.state.IsStreaming.Set(true)
	c.state.Error.Set("")

	// Add placeholder assistant message
	c.state.AddMessage(Message{
		Role:      "assistant",
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFn = cancel

	tokenCh := make(chan string, 100)
	startTime := time.Now()

	go func() {
		msgs := c.state.Messages.Get()
		// Exclude the last (empty assistant) message
		chatMsgs := msgs[:len(msgs)-1]

		err := provider.Chat(ctx, chatMsgs, ChatOpts{
			Model:        c.state.Model.Get(),
			Temperature:  c.state.Temperature.Get(),
			SystemPrompt: c.state.SystemPrompt.Get(),
		}, tokenCh)

		if err != nil && err != context.Canceled {
			c.state.Error.Set(err.Error())
		}
	}()

	// Process tokens in a separate goroutine that sends to our channel
	go func() {
		var content string
		for token := range tokenCh {
			content += token
			c.tokenCh <- content
		}
		// Signal done
		duration := time.Since(startTime)
		c.tokenCh <- "DONE:" + content + "|" + duration.String()
	}()
}

func (c *chatApp) retryLast() {
	msgs := c.state.Messages.Get()
	if len(msgs) < 2 {
		return
	}
	// Remove last assistant message
	c.state.Messages.Set(msgs[:len(msgs)-1])
	c.startStreaming()
}

func (c *chatApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.Watch(c.tokenCh, c.handleToken),
	}
}

func (c *chatApp) handleToken(data string) {
	if len(data) > 5 && data[:5] == "DONE:" {
		// Parse done message
		rest := data[5:]
		// Find duration separator
		for i := len(rest) - 1; i >= 0; i-- {
			if rest[i] == '|' {
				content := rest[:i]
				durStr := rest[i+1:]
				dur, _ := time.ParseDuration(durStr)

				msgs := c.state.Messages.Get()
				if len(msgs) > 0 {
					msgs[len(msgs)-1].Content = content
					msgs[len(msgs)-1].Streaming = false
					msgs[len(msgs)-1].Duration = dur
					c.state.Messages.Set(msgs)
				}
				break
			}
		}
		c.state.IsStreaming.Set(false)
		c.cancelFn = nil
		c.events.Emit(ChatEvent{Type: "done"})
	} else {
		c.state.UpdateLastMessage(data, false)
		c.events.Emit(ChatEvent{Type: "token"})
	}
}

func (c *chatApp) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) {
			if c.state.IsStreaming.Get() {
				c.events.Emit(ChatEvent{Type: "cancel"})
			} else {
				tui.Stop()
			}
		}),
		tui.OnKey(tui.KeyCtrlL, func(ke tui.KeyEvent) {
			c.state.ClearMessages()
		}),
		tui.OnRune('?', func(ke tui.KeyEvent) {
			c.showHelp.Set(!c.showHelp.Get())
		}),
	}

	// Close help on any key when shown
	if c.showHelp.Get() {
		km = append(km, tui.OnRunesStop(func(ke tui.KeyEvent) {
			c.showHelp.Set(false)
		}))
		km = append(km, tui.OnKeyStop(tui.KeyEscape, func(ke tui.KeyEvent) {
			c.showHelp.Set(false)
		}))
	}

	return km
}

templ (c *chatApp) Render() {
	<div class="flex-col h-full">
		@Header(c.state)
		@if c.showHelp.Get() {
			@HelpOverlay()
		} @else {
			@MessageList(c.state, c.events)
		}
		@if c.state.Error.Get() != "" {
			<div class="border-rounded border-red p-1 m-1">
				<span class="text-red">{" Error: " + c.state.Error.Get()}</span>
			</div>
		}
		@InputBar(c.state, c.events)
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate app.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add ChatApp root component with streaming"
```

---

## Task 10: Update main.go

**Files:**
- Modify: `examples/ai-chat/main.go`

**Step 1: Update main.go to use ChatApp**

Replace `examples/ai-chat/main.go`:
```go
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate .

func main() {
	// Detect available providers
	providers := DetectProviders()
	if len(providers) == 0 {
		fmt.Fprintln(os.Stderr, "No providers available. Set one of:")
		fmt.Fprintln(os.Stderr, "  OPENAI_API_KEY")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_API_KEY")
		fmt.Fprintln(os.Stderr, "  Or have Ollama running locally")
		os.Exit(1)
	}

	// Initialize state
	state := NewAppState()

	// Set available providers and select first
	for name := range providers {
		state.AvailableProviders = append(state.AvailableProviders, name)
	}
	if len(state.AvailableProviders) > 0 {
		state.Provider.Set(state.AvailableProviders[0])
		models := state.ProviderModels[state.AvailableProviders[0]]
		if len(models) > 0 {
			state.Model.Set(models[0])
		}
	}

	// Create app
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Set root component
	app.SetRoot(ChatApp(state, providers))

	// Run
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Generate all components and build**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate . && go build
```

**Step 3: Test with a provider**

```bash
cd examples/ai-chat && OPENAI_API_KEY=your-key ./ai-chat
```

Expected: Chat interface appears, can send messages and receive streaming responses

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): wire up main.go with provider detection"
```

---

## Task 11: Settings Screen - Types and Main

**Files:**
- Create: `examples/ai-chat/settings/main.go`

**Step 1: Create settings/main.go**

```bash
mkdir -p examples/ai-chat/settings
```

Create `examples/ai-chat/settings/main.go`:
```go
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
```

**Step 2: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 3: Commit**

```bash
gcommit -m "feat(ai-chat): add settings screen entry point"
```

---

## Task 12: Settings Components

**Files:**
- Create: `examples/ai-chat/settings/settings.gsx`

**Step 1: Create settings.gsx**

Create `examples/ai-chat/settings/settings.gsx`:
```gsx
package settings

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type settingsApp struct {
	state    *SettingsState
	saveBtn  *tui.Ref
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
	s.state.FocusedSection.Set((s.state.FocusedSection.Get() + 1) % 4)
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
	idx = (idx + dir + len(providers)) % len(providers)
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
	idx = (idx + dir + len(models)) % len(models)
	s.state.Model.Set(models[idx])
}

func (s *settingsApp) adjustTemp(delta float64) {
	t := s.state.Temperature.Get() + delta
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	s.state.Temperature.Set(t)
}

func (s *settingsApp) sectionClass(section int) string {
	if s.state.FocusedSection.Get() == section {
		return "border-rounded border-cyan p-1"
	}
	return "border-rounded p-1"
}

func (s *settingsApp) tempBar() string {
	t := s.state.Temperature.Get()
	filled := int(t * 30)
	bar := ""
	for i := 0; i < 30; i++ {
		if i < filled {
			bar += ""
		} else if i == filled {
			bar += ""
		} else {
			bar += ""
		}
	}
	return bar
}

templ (s *settingsApp) Render() {
	<div class="flex-col h-full p-2 gap-2">
		<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifyCenter} align={tui.AlignCenter}>
			<span class="text-gradient-cyan-magenta font-bold">{"  Settings"}</span>
		</div>

		<div class={s.sectionClass(0)}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Provider"}</span>
				<div class="flex gap-2">
					@for _, p := range s.state.AvailableProviders {
						@if p == s.state.Provider.Get() {
							<span class="text-cyan font-bold">{" " + p}</span>
						} @else {
							<span class="font-dim">{" " + p}</span>
						}
					}
				</div>
			</div>
		</div>

		<div class={s.sectionClass(1)}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Model"}</span>
				<div class="flex gap-2">
					@for _, m := range s.state.ProviderModels[s.state.Provider.Get()] {
						@if m == s.state.Model.Get() {
							<span class="text-cyan font-bold">{" " + m}</span>
						} @else {
							<span class="font-dim">{" " + m}</span>
						}
					}
				</div>
			</div>
		</div>

		<div class={s.sectionClass(2)}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Temperature"}</span>
				<div class="flex gap-2 items-center">
					<span class="text-white">{s.tempBar()}</span>
					<span class="text-cyan">{fmt.Sprintf("%.1f", s.state.Temperature.Get())}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">{"  creative"}</span>
					<span class="font-dim">{"precise "}</span>
				</div>
			</div>
		</div>

		<div class={s.sectionClass(3)} flexGrow={1}>
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
			<span class="font-dim">{"Tab: navigate  /: select  Enter: save  Esc: cancel"}</span>
		</div>
	</div>
}
```

**Step 2: Generate Go code**

```bash
cd examples/ai-chat/settings && go run ../../../cmd/tui generate settings.gsx
```

**Step 3: Verify compilation**

```bash
cd examples/ai-chat && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): add Settings screen components"
```

---

## Task 13: Wire Settings to Main App

**Files:**
- Modify: `examples/ai-chat/app.gsx`

**Step 1: Add settings trigger to ChatApp KeyMap**

Update the KeyMap in `app.gsx` to add Ctrl+, handler:

In `examples/ai-chat/app.gsx`, add to the KeyMap function after the existing keybindings:
```go
// Add this import at top
import (
	"github.com/grindlemire/go-tui/examples/ai-chat/settings"
)

// Add to KeyMap function
tui.OnKey(tui.KeyCtrlBackslash, func(ke tui.KeyEvent) { // Ctrl+\ as fallback for Ctrl+,
    c.openSettings()
}),
```

Add the openSettings method:
```go
func (c *chatApp) openSettings() {
	result := settings.Show(
		c.state.Provider.Get(),
		c.state.Model.Get(),
		c.state.Temperature.Get(),
		c.state.SystemPrompt.Get(),
		c.state.AvailableProviders,
		c.state.ProviderModels,
	)
	if result.Saved {
		c.state.Provider.Set(result.Provider)
		c.state.Model.Set(result.Model)
		c.state.Temperature.Set(result.Temperature)
		c.state.SystemPrompt.Set(result.SystemPrompt)
	}
}
```

**Step 2: Update app.gsx with full settings integration**

Replace the chatApp struct and methods to include settings.

**Step 3: Regenerate and build**

```bash
cd examples/ai-chat && go run ../../cmd/tui generate . && go build
```

**Step 4: Commit**

```bash
gcommit -m "feat(ai-chat): wire settings screen to main app"
```

---

## Task 14: Final Polish and Testing

**Step 1: Test all providers**

```bash
# Test OpenAI
OPENAI_API_KEY=your-key ./ai-chat

# Test Anthropic
ANTHROPIC_API_KEY=your-key ./ai-chat

# Test Ollama (if running)
./ai-chat
```

**Step 2: Test all keyboard shortcuts**

- `?` - Help overlay toggle
- `Ctrl+L` - Clear conversation
- `Ctrl+C` - Cancel streaming / quit
- `j/k` - Navigate messages
- `c` - Copy message
- `r` - Retry response
- `Ctrl+\` - Open settings

**Step 3: Test streaming**

Send a message and verify:
- Yellow border on input while streaming
- Blinking cursor on response
- Response time shown after completion
- Scroll follows response

**Step 4: Commit final polish**

```bash
gcommit -m "feat(ai-chat): complete polished AI chat example"
```

---

## Summary

This implementation plan creates a full-featured AI chat application with:

1. **Multi-provider support** (OpenAI, Anthropic, Ollama)
2. **Streaming responses** via channels and Watchers
3. **Component composition** (Header, MessageList, Message, InputBar, HelpOverlay)
4. **Shared state** via `*tui.State[T]` pointers
5. **Event bus** for cross-component communication
6. **References** for scroll control and click handling
7. **Alternate buffer** settings screen
8. **Beautiful styling** with gradients and rounded borders
9. **Comprehensive keyboard navigation**

Total: 14 tasks, ~45 minutes estimated implementation time.
