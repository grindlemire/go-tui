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
	// Find and invoke the Escape binding
	fired := false
	for _, b := range km {
		if b.Pattern.Key == KeyEscape {
			b.Handler(KeyEvent{Key: KeyEscape})
			fired = true
			break
		}
	}
	if !fired {
		t.Fatal("no Escape binding found in KeyMap")
	}
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
	// Should still have Tab bindings (from trapFocus) but no Escape
	for _, b := range km {
		if b.Pattern.Key == KeyEscape {
			t.Error("expected no Escape binding when closeOnEscape is false")
		}
	}
}

func TestModal_KeyMap_TabFocusCycling(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	km := m.KeyMap()
	hasTab := false
	hasShiftTab := false
	for _, b := range km {
		if b.Pattern.Key == KeyTab && b.Pattern.Mod == 0 {
			hasTab = true
		}
		if b.Pattern.Key == KeyTab && b.Pattern.Mod == ModShift {
			hasShiftTab = true
		}
	}
	if !hasTab {
		t.Error("expected Tab binding when trapFocus is true")
	}
	if !hasShiftTab {
		t.Error("expected Shift+Tab binding when trapFocus is true")
	}
}

func TestModal_KeyMap_NoTabWhenTrapFocusDisabled(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open), WithModalTrapFocus(false))
	m.BindApp(testApp)

	km := m.KeyMap()
	for _, b := range km {
		if b.Pattern.Key == KeyTab {
			t.Error("expected no Tab binding when trapFocus is false")
		}
	}
}

func TestModal_HandleMouse_BackdropClick(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	// Render the modal to set up the element and overlay
	el := m.Render(testApp)
	// Trigger layout by rendering into a buffer
	buf := NewBuffer(80, 24)
	el.Render(buf, 80, 24)

	// Click on the overlay element itself (backdrop area, no children)
	consumed := m.HandleMouse(MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      0,
		Y:      0,
	})

	if !consumed {
		t.Error("expected backdrop click to be consumed")
	}
	if open.Get() {
		t.Error("expected open to be false after backdrop click")
	}

	testApp.clearOverlays()
}

func TestModal_HandleMouse_BackdropClickDisabled(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open), WithModalCloseOnBackdropClick(false))
	m.BindApp(testApp)

	el := m.Render(testApp)
	buf := NewBuffer(80, 24)
	el.Render(buf, 80, 24)

	consumed := m.HandleMouse(MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      0,
		Y:      0,
	})

	if !consumed {
		t.Error("expected backdrop click to be consumed even when close is disabled")
	}
	if !open.Get() {
		t.Error("expected open to remain true when backdrop click is disabled")
	}

	testApp.clearOverlays()
}

func TestModal_HandleMouse_ChildOnActivate(t *testing.T) {
	open := NewState(true)
	activated := false
	m := NewModal(
		WithModalOpen(open),
		WithModalElementOptions(WithDirection(Column)),
	)
	m.BindApp(testApp)

	el := m.Render(testApp)
	// Add a child button with onActivate and explicit size
	btn := New(WithOnActivate(func() { activated = true }), WithWidth(10), WithHeight(1))
	el.AddChild(btn)

	// Trigger layout
	buf := NewBuffer(80, 24)
	el.Render(buf, 80, 24)

	// Click within the button's rendered bounds
	btnRect := btn.Rect()
	consumed := m.HandleMouse(MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      btnRect.X,
		Y:      btnRect.Y,
	})

	if !consumed {
		t.Error("expected child click to be consumed")
	}
	if !activated {
		t.Error("expected onActivate to be called")
	}

	testApp.clearOverlays()
}

func TestModal_KeyMap_EnterActivatesFocused(t *testing.T) {
	open := NewState(true)
	activated := false

	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	// Create a focusable element with onActivate
	btn := New(WithOnActivate(func() { activated = true }), WithFocusable(true))
	testApp.focus = newFocusManager()
	testApp.focus.Register(btn)
	testApp.focus.Next() // focus the button

	km := m.KeyMap()
	// Find the Enter binding
	for _, b := range km {
		if b.Pattern.Key == KeyEnter {
			b.Handler(KeyEvent{Key: KeyEnter})
			break
		}
	}

	if !activated {
		t.Error("expected Enter to trigger onActivate on focused element")
	}
}

func TestModal_HandleMouse_ClosedNoOp(t *testing.T) {
	open := NewState(false)
	m := NewModal(WithModalOpen(open))
	m.BindApp(testApp)

	consumed := m.HandleMouse(MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      5,
		Y:      5,
	})

	if consumed {
		t.Error("expected mouse event to not be consumed when modal is closed")
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
