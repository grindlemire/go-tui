package tui

import "testing"

func TestModal_NewModal_Defaults(t *testing.T) {
	m := NewModal()

	if m.backdrop != "dim" {
		t.Errorf("expected default backdrop 'dim', got %q", m.backdrop)
	}
	if !m.closeOnEscape {
		t.Error("expected closeOnEscape default true")
	}
	if !m.closeOnBackdrop {
		t.Error("expected closeOnBackdrop default true")
	}
	if !m.trapFocus {
		t.Error("expected trapFocus default true")
	}
}

func TestModal_Render_Closed(t *testing.T) {
	open := NewState(false)
	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	el := m.Render(testApp)
	if !el.IsOverlay() {
		t.Error("expected overlay flag on element")
	}
	if !el.hidden {
		t.Error("expected hidden element when modal is closed")
	}
}

func TestModal_Render_Open(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	el := m.Render(testApp)
	if !el.IsOverlay() {
		t.Error("expected overlay flag on element")
	}
	if el.hidden {
		t.Error("expected visible element when modal is open")
	}
	if len(testApp.overlays) == 0 {
		t.Error("expected overlay to be registered")
	}
	// Clean up
	testApp.clearOverlays()
}

func TestModal_KeyMap_Escape(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open))

	km := m.KeyMap()
	if len(km) == 0 {
		t.Fatal("expected non-empty KeyMap when open")
	}
	// Simulate Escape
	km[0].Handler(KeyEvent{Key: KeyEscape})
	if open.Get() {
		t.Error("expected open to be false after Escape")
	}
}

func TestModal_KeyMap_Closed(t *testing.T) {
	open := NewState(false)
	m := NewModal(WithModalOpen(open))

	km := m.KeyMap()
	if km != nil {
		t.Error("expected nil KeyMap when closed")
	}
}

func TestModal_KeyMap_EscapeDisabled(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open), WithModalCloseOnEscape(false))

	km := m.KeyMap()
	if km != nil {
		t.Error("expected nil KeyMap when closeOnEscape is false")
	}
}

func TestModal_Options(t *testing.T) {
	type tc struct {
		opts     []ModalOption
		backdrop string
		escape   bool
		click    bool
		focus    bool
	}

	tests := map[string]tc{
		"all defaults": {
			opts:     nil,
			backdrop: "dim",
			escape:   true,
			click:    true,
			focus:    true,
		},
		"custom backdrop": {
			opts:     []ModalOption{WithModalBackdrop("blank")},
			backdrop: "blank",
			escape:   true,
			click:    true,
			focus:    true,
		},
		"no escape": {
			opts:     []ModalOption{WithModalCloseOnEscape(false)},
			backdrop: "dim",
			escape:   false,
			click:    true,
			focus:    true,
		},
		"no backdrop click": {
			opts:     []ModalOption{WithModalCloseOnBackdropClick(false)},
			backdrop: "dim",
			escape:   true,
			click:    false,
			focus:    true,
		},
		"no focus trap": {
			opts:     []ModalOption{WithModalTrapFocus(false)},
			backdrop: "dim",
			escape:   true,
			click:    true,
			focus:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewModal(tt.opts...)
			if m.backdrop != tt.backdrop {
				t.Errorf("backdrop: got %q, want %q", m.backdrop, tt.backdrop)
			}
			if m.closeOnEscape != tt.escape {
				t.Errorf("closeOnEscape: got %v, want %v", m.closeOnEscape, tt.escape)
			}
			if m.closeOnBackdrop != tt.click {
				t.Errorf("closeOnBackdrop: got %v, want %v", m.closeOnBackdrop, tt.click)
			}
			if m.trapFocus != tt.focus {
				t.Errorf("trapFocus: got %v, want %v", m.trapFocus, tt.focus)
			}
		})
	}
}
