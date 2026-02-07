package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

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

// --- Fake Provider (for testing/demo) ---

type FakeProvider struct{}

func NewFakeProvider() *FakeProvider {
	return &FakeProvider{}
}

func (p *FakeProvider) Name() string { return "fake" }

// Lorem ipsum words for generating fake responses
var loremWords = strings.Split(
	"Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum",
	" ",
)

func (p *FakeProvider) Chat(ctx context.Context, messages []Message, opts ChatOpts, tokenCh chan<- string) error {
	defer close(tokenCh)

	// Generate 50-200 words of lorem ipsum
	wordCount := 50 + rand.Intn(151)

	// Total duration 1-10 seconds
	totalDuration := time.Duration(1+rand.Intn(10)) * time.Second
	delayPerWord := totalDuration / time.Duration(wordCount)

	for i := 0; i < wordCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		word := loremWords[rand.Intn(len(loremWords))]

		// Add space before word (except first)
		if i > 0 {
			word = " " + word
		}

		// Add punctuation occasionally
		if i > 0 && rand.Float32() < 0.15 {
			punctuation := []string{".", ",", "!", "?"}
			word = punctuation[rand.Intn(len(punctuation))] + word
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case tokenCh <- word:
		}

		time.Sleep(delayPerWord)
	}

	// End with a period
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tokenCh <- ".":
	}

	return nil
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

	// Fake provider is always available (for testing/demo)
	providers["fake"] = NewFakeProvider()

	return providers
}
