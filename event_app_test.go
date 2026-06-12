package tui

import "testing"

func TestMouseEvent_App(t *testing.T) {
	app := &App{}

	type tc struct {
		event MouseEvent
		want  *App
	}

	tests := map[string]tc{
		"returns the dispatching app": {
			event: MouseEvent{Button: MouseLeft, Action: MousePress, X: 3, Y: 4, app: app},
			want:  app,
		},
		"zero value returns nil": {
			event: MouseEvent{},
			want:  nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.event.App(); got != tt.want {
				t.Errorf("MouseEvent.App() = %p, want %p", got, tt.want)
			}
		})
	}
}

func TestKeyEvent_App(t *testing.T) {
	app := &App{}

	type tc struct {
		event KeyEvent
		want  *App
	}

	tests := map[string]tc{
		"returns the dispatching app": {
			event: KeyEvent{Key: KeyEnter, app: app},
			want:  app,
		},
		"zero value returns nil": {
			event: KeyEvent{},
			want:  nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.event.App(); got != tt.want {
				t.Errorf("KeyEvent.App() = %p, want %p", got, tt.want)
			}
		})
	}
}
