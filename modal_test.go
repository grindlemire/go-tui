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

// testModalRoot is a minimal root component that wraps a Modal for testing.
type testModalRoot struct {
	modal *Modal
}

func (r *testModalRoot) Render(app *App) *Element {
	root := New()
	el := r.modal.Render(app)
	root.AddChild(el)
	return root
}

// newTestApp creates a lightweight App with a mock terminal and buffer for modal tests.
func newTestApp(width, height int) *App {
	return &App{
		terminal:    NewMockTerminal(width, height),
		buffer:      NewBuffer(width, height),
		stopCh:      make(chan struct{}),
		eventQueue:  make(chan func(), 256),
		updateQueue: make(chan func(), 256),
		focus:       newFocusManager(),
		mounts:      newMountState(),
		batch:       newBatchContext(),
	}
}

func TestModal_RenderFull_RepopulatesOverlays(t *testing.T) {
	open := NewState(true)
	modal := NewModal(WithModalOpen(open))
	rootComp := &testModalRoot{modal: modal}

	app := newTestApp(80, 24)
	modal.BindApp(app)
	app.rootComponent = rootComp

	// First render populates overlays
	app.Render()
	if len(app.overlays) == 0 {
		t.Fatal("expected overlay after initial Render()")
	}

	// RenderFull should re-render the component tree and repopulate overlays
	app.RenderFull()
	if len(app.overlays) == 0 {
		t.Error("RenderFull() cleared overlays without re-registering them; open modal vanishes on full redraw")
	}
}

func TestModal_RenderFull_NeedsFocusInit(t *testing.T) {
	open := NewState(false)
	modal := NewModal(WithModalOpen(open))
	rootComp := &testModalRoot{modal: modal}

	app := newTestApp(80, 24)
	modal.BindApp(app)
	app.rootComponent = rootComp

	// Initial render with modal closed
	app.Render()
	if len(app.overlays) != 0 {
		t.Fatal("expected no overlays when modal is closed")
	}

	// Open the modal; next render should set needsFocusInit
	open.Set(true)

	// Use RenderFull to verify it processes needsFocusInit
	app.RenderFull()
	if len(app.overlays) == 0 {
		t.Fatal("expected overlay after opening modal")
	}
	for _, ov := range app.overlays {
		if ov.needsFocusInit {
			t.Error("RenderFull() did not process needsFocusInit; focus won't auto-enter the modal on full redraw")
		}
	}
}

func TestModal_KeyMap_InlineMode(t *testing.T) {
	open := NewState(true)
	m := NewModal(WithModalOpen(open))

	app := newTestApp(80, 24)
	app.inlineHeight = 10
	app.inAlternateScreen = false
	m.BindApp(app)

	// In inline mode, registerOverlay silently skips the overlay
	m.Render(app)
	if len(app.overlays) != 0 {
		t.Fatal("expected no overlay in inline mode")
	}

	// KeyMap should return nil since the modal is not rendered
	km := m.KeyMap()
	if km != nil {
		t.Errorf("expected nil KeyMap in inline mode, got %d bindings", len(km))
	}

	// After entering alternate screen, KeyMap should return bindings
	app.inAlternateScreen = true
	km = m.KeyMap()
	if km == nil {
		t.Error("expected non-nil KeyMap in alternate screen mode")
	}
}

func TestModal_WithModalBackdrop_InvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid backdrop value")
		}
	}()
	WithModalBackdrop("typo")
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
