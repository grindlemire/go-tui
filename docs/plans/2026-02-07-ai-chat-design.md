# AI Chat Example Design

A polished, full-featured AI chat application showcasing go-tui's capabilities with langchaingo integration.

## Features

- **Multi-provider support**: OpenAI, Anthropic, Ollama (runtime configurable)
- **Streaming responses**: Tokens appear as generated with typing indicator
- **Rich UI**: Gradient text, rounded borders, response times, token counts
- **Settings screen**: Full alternate-buffer settings with provider/model/temperature/system prompt
- **Keyboard-driven**: Vim-style navigation, comprehensive shortcuts

## Visual Design

### Main Chat Screen

```
╭─────────────────────────────────────────────────────────────────────────╮
│  ✦ AI Chat                          gpt-4 │ 1,234 tokens │ Ctrl+? help  │
╰─────────────────────────────────────────────────────────────────────────╯

   ╭──────────────────────────────────────────────────────────────────╮
   │  ● Assistant                                                     │
   │                                                                  │
   │  Hello! I'm your AI assistant. How can I help you today?        │
   ╰──────────────────────────────────────────────────────────────────╯

   ╭──────────────────────────────────────────────────────────────────╮
   │  ○ You                                                           │
   │                                                                  │
   │  Explain goroutines in Go                                        │
   ╰──────────────────────────────────────────────────────────────────╯

   ╭──────────────────────────────────────────────────────────────────╮
   │  ● Assistant                                        1.2s  ↻  📋  │
   │                                                                  │
   │  Goroutines are lightweight threads managed by the Go runtime.   │
   │  They're incredibly efficient - you can spawn thousands with     │
   │  minimal overhead...█                                            │
   ╰──────────────────────────────────────────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────╮
│  › Explain how channels work with goroutines...                     ⏎   │
╰─────────────────────────────────────────────────────────────────────────╯
```

### Settings Screen (Alternate Buffer)

```
╭─────────────────────────────────────────────────────────────────────────╮
│                         ✦ Settings                                      │
╰─────────────────────────────────────────────────────────────────────────╯

  ╭─ Provider ────────────────────────────────────────────────────────────╮
  │                                                                       │
  │    ● OpenAI        ○ Anthropic        ○ Ollama                       │
  │                                                                       │
  ╰───────────────────────────────────────────────────────────────────────╯

  ╭─ Model ───────────────────────────────────────────────────────────────╮
  │                                                                       │
  │    ● gpt-4         ○ gpt-4-turbo      ○ gpt-3.5-turbo               │
  │                                                                       │
  ╰───────────────────────────────────────────────────────────────────────╯

  ╭─ Temperature ─────────────────────────────────────────────────────────╮
  │                                                                       │
  │    ━━━━━━━━━━━━━━━━━━━━━●━━━━━━━━━━━━━━━━━━━━━━━━━━━  0.7            │
  │    ← creative                                     precise →          │
  │                                                                       │
  ╰───────────────────────────────────────────────────────────────────────╯

  ╭─ System Prompt ───────────────────────────────────────────────────────╮
  │                                                                       │
  │  You are a helpful, concise assistant. Answer questions clearly      │
  │  and provide code examples when relevant.                            │
  │                                                                       │
  ╰───────────────────────────────────────────────────────────────────────╯

                              [Save]  [Cancel]

                         Esc to cancel • Enter to save
```

## Architecture

### Component Tree

```
ChatApp (root)
├── Header
│   ├── ModelDisplay
│   ├── TokenCounter
│   └── HelpHint
├── MessageList (scrollable)
│   └── Message (repeated)
│       ├── MessageHeader (role, time, buttons)
│       └── MessageContent (text, streaming cursor)
├── InputBar
│   └── TextInput
└── HelpOverlay (conditional)

SettingsApp (alternate buffer)
├── Header
├── ProviderSelect
├── ModelSelect
├── TemperatureSlider
├── SystemPromptEditor
└── ActionButtons
```

### State Management

**Shared State (AppState):**
```go
type AppState struct {
    // Provider config
    Provider     *State[string]      // "openai" | "anthropic" | "ollama"
    Model        *State[string]      // "gpt-4", "claude-3", etc.
    Temperature  *State[float64]     // 0.0 - 1.0
    SystemPrompt *State[string]

    // Conversation
    Messages     *State[[]Message]

    // UI state
    TotalTokens  *State[int]
    IsStreaming  *State[bool]
}

type Message struct {
    Role      string    // "user" | "assistant"
    Content   string
    Tokens    int
    Duration  time.Duration
    Timestamp time.Time
}
```

**Event Bus:**
```go
type ChatEvent struct {
    Type    string  // "token" | "done" | "error" | "retry" | "copy"
    Payload any
}

events := tui.NewEvents[ChatEvent]()
```

**Data Flow:**
```
User types → InputBar.Submit()
                ↓
         AppState.Messages updated (user msg)
                ↓
         events.Emit({Type: "token", ...}) ←── streaming goroutine
                ↓
         MessageList receives tokens via subscription
                ↓
         events.Emit({Type: "done"})
                ↓
         AppState.Messages updated (final assistant msg)
         AppState.TotalTokens updated
```

## LangChainGo Integration

### Provider Interface

```go
type Provider interface {
    Name() string
    Models() []string
    Chat(ctx context.Context, messages []Message, opts ChatOpts) (<-chan string, error)
}

type ChatOpts struct {
    Model        string
    Temperature  float64
    SystemPrompt string
}
```

### Supported Providers

| Provider | Env Var | Default Models |
|----------|---------|----------------|
| OpenAI | `OPENAI_API_KEY` | gpt-4, gpt-4-turbo, gpt-3.5-turbo |
| Anthropic | `ANTHROPIC_API_KEY` | claude-3-opus, claude-3-sonnet, claude-3-haiku |
| Ollama | `OLLAMA_HOST` (optional) | llama2, mistral, codellama |

## Keyboard Shortcuts

### Global (ChatApp)

| Key | Action |
|-----|--------|
| `Ctrl+,` | Open settings (alternate buffer) |
| `Ctrl+?` or `?` | Toggle help overlay |
| `Ctrl+C` | Cancel streaming / Exit app |
| `Ctrl+L` | Clear conversation |
| `Ctrl+N` | New conversation (reset) |
| `Esc` | Cancel streaming if active |

### Message List

| Key | Action |
|-----|--------|
| `j` / `↓` | Next message |
| `k` / `↑` | Previous message |
| `g` | Jump to first message |
| `G` | Jump to last message |
| `c` | Copy focused message |
| `r` | Retry focused message (if assistant) |

### Input Bar

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` | Edit last user message (when input empty) |

### Settings Screen

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Navigate sections |
| `←` / `→` | Change selection / adjust slider |
| `Enter` | Save and return |
| `Esc` | Cancel and return |

## File Structure

```
examples/ai-chat/
├── main.go                 # Entry point, provider detection, app setup
├── state.go                # AppState, Message types, ChatEvent
├── providers.go            # Provider interface + implementations
│
├── app.gsx                 # ChatApp root component
├── header.gsx              # Header with model display, token count, help
├── message_list.gsx        # Scrollable message container
├── message.gsx             # Individual message with actions
├── input_bar.gsx           # Text input with submit
├── help_overlay.gsx        # Keyboard shortcuts help (conditional)
│
├── settings/
│   ├── main.go             # Settings app entry (alternate buffer)
│   ├── settings.gsx        # Settings root component
│   ├── provider_select.gsx # Radio group for providers
│   ├── model_select.gsx    # Radio group for models
│   ├── temp_slider.gsx     # Temperature slider
│   └── prompt_editor.gsx   # System prompt text area
│
└── go.mod                  # Module with langchaingo dependency
```

## Component Details

| Component | Local State | Shared State | Events |
|-----------|-------------|--------------|--------|
| `ChatApp` | helpVisible | all AppState | subscribes to all |
| `Header` | - | Provider, Model, TotalTokens | - |
| `MessageList` | scrollY, focusedIdx | Messages, IsStreaming | subscribes to token/done |
| `Message` | hovered | - | emits copy/retry |
| `InputBar` | inputText | IsStreaming | emits submit |
| `Settings*` | local form state | reads/writes AppState | - |

## References

| Component | Refs | Purpose |
|-----------|------|---------|
| `MessageList` | `content` | Scroll control |
| `Message` | `copyBtn`, `retryBtn` | Click handling |
| `InputBar` | `input` | Focus management |
| `TempSlider` | `track` | Click position calculation |
