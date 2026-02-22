package main

import (
	"strings"
	"testing"

	tui "github.com/grindlemire/go-tui"
)

func TestCounter_KeyMap(t *testing.T) {
	type tc struct {
		key       tui.KeyEvent
		wantCount int
	}

	tests := map[string]tc{
		"increment with +": {
			key:       tui.KeyEvent{Key: tui.KeyRune, Rune: '+'},
			wantCount: 1,
		},
		"decrement with -": {
			key:       tui.KeyEvent{Key: tui.KeyRune, Rune: '-'},
			wantCount: -1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCounter()
			km := c.KeyMap()

			// Find the matching binding and call its handler
			for _, binding := range km {
				if binding.Pattern.AnyRune && tt.key.IsRune() {
					binding.Handler(tt.key)
					break
				}
				if binding.Pattern.Rune == tt.key.Rune && tt.key.IsRune() {
					binding.Handler(tt.key)
					break
				}
				if binding.Pattern.Key == tt.key.Key && !tt.key.IsRune() {
					binding.Handler(tt.key)
					break
				}
			}

			if c.count.Get() != tt.wantCount {
				t.Errorf("count = %d, want %d", c.count.Get(), tt.wantCount)
			}
		})
	}
}

func TestCounter_MultipleKeyPresses(t *testing.T) {
	type tc struct {
		initialCount int
		keyPresses   []rune
		wantCount    int
	}

	tests := map[string]tc{
		"no keys pressed": {
			initialCount: 0,
			keyPresses:   nil,
			wantCount:    0,
		},
		"three increments": {
			initialCount: 0,
			keyPresses:   []rune{'+', '+', '+'},
			wantCount:    3,
		},
		"increment then decrement": {
			initialCount: 5,
			keyPresses:   []rune{'+', '-'},
			wantCount:    5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCounter()
			c.count.Set(tt.initialCount)

			km := c.KeyMap()
			for _, r := range tt.keyPresses {
				event := tui.KeyEvent{Key: tui.KeyRune, Rune: r}
				for _, binding := range km {
					if binding.Pattern.Rune == r {
						binding.Handler(event)
						break
					}
				}
			}

			if c.count.Get() != tt.wantCount {
				t.Errorf("count = %d, want %d", c.count.Get(), tt.wantCount)
			}
		})
	}
}

func TestPanel_RendersBorder(t *testing.T) {
	root := tui.New(
		tui.WithSize(80, 24),
		tui.WithDirection(tui.Column),
	)

	panel := tui.New(
		tui.WithSize(20, 5),
		tui.WithBorder(tui.BorderSingle),
	)
	root.AddChild(panel)

	buf := tui.NewBuffer(80, 24)
	root.Render(buf, 80, 24)

	rect := panel.Rect()
	if rect.Width != 20 || rect.Height != 5 {
		t.Errorf("panel size = %dx%d, want 20x5", rect.Width, rect.Height)
	}

	cell := buf.Cell(rect.X, rect.Y)
	if cell.Rune != '┌' {
		t.Errorf("top-left = %q, want '┌'", cell.Rune)
	}
}

func TestGreeting_ShowsMessage(t *testing.T) {
	buf := tui.NewBuffer(30, 5)
	term := tui.NewMockTerminal(30, 5)
	style := tui.NewStyle()

	buf.SetString(2, 1, "Hello, World!", style)
	tui.Render(term, buf)

	output := term.StringTrimmed()
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("expected 'Hello, World!' in output:\n%s", output)
	}
}

func TestLayout_AdaptsToResize(t *testing.T) {
	type tc struct {
		width      int
		height     int
		wantPanelW int
	}

	tests := map[string]tc{
		"narrow terminal": {
			width:      40,
			height:     24,
			wantPanelW: 40,
		},
		"wide terminal": {
			width:      120,
			height:     24,
			wantPanelW: 120,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := tui.New(
				tui.WithDirection(tui.Column),
			)

			panel := tui.New(
				tui.WithFlexGrow(1),
				tui.WithHeight(10),
			)
			root.AddChild(panel)

			buf := tui.NewBuffer(tt.width, tt.height)
			root.Render(buf, tt.width, tt.height)

			rect := panel.Rect()
			if rect.Width != tt.wantPanelW {
				t.Errorf("panel width = %d, want %d", rect.Width, tt.wantPanelW)
			}
		})
	}
}
