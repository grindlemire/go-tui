package main

import "encoding/json"

// streamEventType represents the kinds of events we extract from claude's stream-json output.
type streamEventType int

const (
	eventText     streamEventType = iota // text delta to display
	eventToolUse                         // tool invocation summary
	eventDone                            // response complete
	eventError                           // subprocess error
)

// streamEvent is a parsed event from claude's stream-json output.
type streamEvent struct {
	Type streamEventType
	Text string // for eventText: the text fragment; for eventToolUse: tool name; for eventError: error message
}

// streamParser tracks state across lines to compute text deltas.
// Claude CLI's stream-json (with --include-partial-messages) emits assistant
// events containing the full message so far. We diff against previous state
// to extract only new text.
type streamParser struct {
	seenTextLen map[int]int  // content block index → runes already emitted
	seenTools   map[int]bool // content block index → tool_use already emitted
}

func newStreamParser() *streamParser {
	return &streamParser{
		seenTextLen: make(map[int]int),
		seenTools:   make(map[int]bool),
	}
}

// parseLine parses a single line of claude stream-json output.
// Returns zero or more events (text deltas, tool starts, done).
func (p *streamParser) parseLine(line []byte) []streamEvent {
	if len(line) == 0 {
		return nil
	}

	var raw rawEvent
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil
	}

	switch raw.Type {
	case "result":
		return []streamEvent{{Type: eventDone}}

	case "assistant":
		return p.parseAssistant(raw.Message)

	default:
		return nil
	}
}

func (p *streamParser) parseAssistant(data json.RawMessage) []streamEvent {
	if len(data) == 0 {
		return nil
	}

	var msg rawMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil
	}

	var events []streamEvent
	for i, block := range msg.Content {
		switch block.Type {
		case "text":
			prevLen := p.seenTextLen[i]
			runes := []rune(block.Text)
			if len(runes) > prevLen {
				delta := string(runes[prevLen:])
				p.seenTextLen[i] = len(runes)
				events = append(events, streamEvent{Type: eventText, Text: delta})
			}

		case "tool_use":
			if !p.seenTools[i] && block.Name != "" {
				p.seenTools[i] = true
				events = append(events, streamEvent{Type: eventToolUse, Text: block.Name})
			}
		}
	}
	return events
}

// rawEvent is the top-level JSON structure from claude stream-json.
type rawEvent struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype,omitempty"`
	Message json.RawMessage `json:"message,omitempty"`
}

// rawMessage is the message object inside an assistant event.
type rawMessage struct {
	Content []rawContentBlock `json:"content"`
}

// rawContentBlock is a content block inside an assistant message.
type rawContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}
