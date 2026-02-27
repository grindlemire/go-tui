package main

import "testing"

func TestStreamParserParseLine(t *testing.T) {
	type tc struct {
		name   string
		lines  []string // feed these lines in order
		want   [][]streamEvent
	}

	tests := []tc{
		{
			name:  "empty line",
			lines: []string{""},
			want:  [][]streamEvent{nil},
		},
		{
			name:  "invalid json",
			lines: []string{"not json"},
			want:  [][]streamEvent{nil},
		},
		{
			name:  "result event",
			lines: []string{`{"type":"result","subtype":"success","cost_usd":0.001,"result":"hello"}`},
			want:  [][]streamEvent{{{Type: eventDone}}},
		},
		{
			name:  "system event ignored",
			lines: []string{`{"type":"system","subtype":"init","cwd":"/tmp"}`},
			want:  [][]streamEvent{nil},
		},
		{
			name:  "rate limit event ignored",
			lines: []string{`{"type":"rate_limit_event","rate_limit_info":{}}`},
			want:  [][]streamEvent{nil},
		},
		{
			name: "assistant text extracted",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}`,
			},
			want: [][]streamEvent{
				{{Type: eventText, Text: "Hello"}},
			},
		},
		{
			name: "assistant text delta across two events",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"text","text":"Hel"}]}}`,
				`{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`,
			},
			want: [][]streamEvent{
				{{Type: eventText, Text: "Hel"}},
				{{Type: eventText, Text: "lo world"}},
			},
		},
		{
			name: "thinking blocks ignored",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"let me think"}]}}`,
			},
			want: [][]streamEvent{nil},
		},
		{
			name: "tool use extracted once",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash"}]}}`,
				`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash"}]}}`,
			},
			want: [][]streamEvent{
				{{Type: eventToolUse, Text: "Bash"}},
				nil,
			},
		},
		{
			name: "mixed content blocks",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"text","text":"I'll run a command"},{"type":"tool_use","name":"Edit"}]}}`,
			},
			want: [][]streamEvent{
				{
					{Type: eventText, Text: "I'll run a command"},
					{Type: eventToolUse, Text: "Edit"},
				},
			},
		},
		{
			name: "no delta when text unchanged",
			lines: []string{
				`{"type":"assistant","message":{"content":[{"type":"text","text":"same"}]}}`,
				`{"type":"assistant","message":{"content":[{"type":"text","text":"same"}]}}`,
			},
			want: [][]streamEvent{
				{{Type: eventText, Text: "same"}},
				nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStreamParser()
			for i, line := range tt.lines {
				got := p.parseLine([]byte(line))
				want := tt.want[i]

				if want == nil {
					if len(got) != 0 {
						t.Fatalf("line %d: expected nil, got %+v", i, got)
					}
					continue
				}

				if len(got) != len(want) {
					t.Fatalf("line %d: got %d events, want %d\ngot:  %+v\nwant: %+v", i, len(got), len(want), got, want)
				}

				for j := range want {
					if got[j].Type != want[j].Type {
						t.Fatalf("line %d event %d: type = %d, want %d", i, j, got[j].Type, want[j].Type)
					}
					if got[j].Text != want[j].Text {
						t.Fatalf("line %d event %d: text = %q, want %q", i, j, got[j].Text, want[j].Text)
					}
				}
			}
		})
	}
}
