package tui

import (
	"testing"
	"time"
)

func TestWithOnSuspend(t *testing.T) {
	called := false
	opt := WithOnSuspend(func() { called = true })

	app := &App{}
	err := opt(app)
	if err != nil {
		t.Fatalf("WithOnSuspend returned error: %v", err)
	}
	if app.onSuspend == nil {
		t.Fatal("expected onSuspend to be set")
	}
	app.onSuspend()
	if !called {
		t.Fatal("expected onSuspend callback to be called")
	}
}

func TestWithOnResume(t *testing.T) {
	called := false
	opt := WithOnResume(func() { called = true })

	app := &App{}
	err := opt(app)
	if err != nil {
		t.Fatalf("WithOnResume returned error: %v", err)
	}
	if app.onResume == nil {
		t.Fatal("expected onResume to be set")
	}
	app.onResume()
	if !called {
		t.Fatal("expected onResume callback to be called")
	}
}

func TestWithInputLatency(t *testing.T) {
	type tc struct {
		latency time.Duration
		wantErr bool
		want    time.Duration
	}

	tests := map[string]tc{
		"zero is rejected": {
			latency: 0,
			wantErr: true,
		},
		"positive duration is set": {
			latency: 50 * time.Millisecond,
			want:    50 * time.Millisecond,
		},
		"blocking sentinel is set": {
			latency: InputLatencyBlocking,
			want:    InputLatencyBlocking,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{}
			err := WithInputLatency(tt.latency)(app)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.inputLatency != tt.want {
				t.Fatalf("inputLatency = %v, want %v", app.inputLatency, tt.want)
			}
		})
	}
}

func TestWithFrameRate(t *testing.T) {
	type tc struct {
		fps     int
		wantErr bool
		want    time.Duration
	}

	tests := map[string]tc{
		"zero is rejected": {
			fps:     0,
			wantErr: true,
		},
		"negative is rejected": {
			fps:     -10,
			wantErr: true,
		},
		"above 240 is rejected": {
			fps:     241,
			wantErr: true,
		},
		"minimum 1 fps": {
			fps:  1,
			want: time.Second,
		},
		"60 fps": {
			fps:  60,
			want: time.Second / 60,
		},
		"maximum 240 fps": {
			fps:  240,
			want: time.Second / 240,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{}
			err := WithFrameRate(tt.fps)(app)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.frameDuration != tt.want {
				t.Fatalf("frameDuration = %v, want %v", app.frameDuration, tt.want)
			}
		})
	}
}

func TestWithEventQueueSize(t *testing.T) {
	type tc struct {
		size    int
		wantErr bool
		want    int
	}

	tests := map[string]tc{
		"zero is rejected": {
			size:    0,
			wantErr: true,
		},
		"negative is rejected": {
			size:    -1,
			wantErr: true,
		},
		"minimum size 1": {
			size: 1,
			want: 1,
		},
		"large size": {
			size: 1024,
			want: 1024,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{}
			err := WithEventQueueSize(tt.size)(app)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.eventQueueSize != tt.want {
				t.Fatalf("eventQueueSize = %d, want %d", app.eventQueueSize, tt.want)
			}
		})
	}
}

func TestWithInlineHeight(t *testing.T) {
	type tc struct {
		rows    int
		wantErr bool
		want    int
	}

	tests := map[string]tc{
		"zero is rejected": {
			rows:    0,
			wantErr: true,
		},
		"negative is rejected": {
			rows:    -3,
			wantErr: true,
		},
		"minimum 1 row": {
			rows: 1,
			want: 1,
		},
		"multiple rows": {
			rows: 10,
			want: 10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{}
			err := WithInlineHeight(tt.rows)(app)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.inlineHeight != tt.want {
				t.Fatalf("inlineHeight = %d, want %d", app.inlineHeight, tt.want)
			}
		})
	}
}

func TestWithInlineStartupMode(t *testing.T) {
	type tc struct {
		mode    InlineStartupMode
		wantErr bool
	}

	tests := map[string]tc{
		"preserve visible": {
			mode: InlineStartupPreserveVisible,
		},
		"fresh viewport": {
			mode: InlineStartupFreshViewport,
		},
		"soft reset": {
			mode: InlineStartupSoftReset,
		},
		"invalid mode is rejected": {
			mode:    InlineStartupMode(99),
			wantErr: true,
		},
		"negative mode is rejected": {
			mode:    InlineStartupMode(-1),
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Pre-set a sentinel so we can verify rejected modes leave the
			// field untouched.
			app := &App{inlineStartupMode: InlineStartupPreserveVisible}
			err := WithInlineStartupMode(tt.mode)(app)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if app.inlineStartupMode != InlineStartupPreserveVisible {
					t.Fatalf("inlineStartupMode mutated to %d on error", app.inlineStartupMode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.inlineStartupMode != tt.mode {
				t.Fatalf("inlineStartupMode = %d, want %d", app.inlineStartupMode, tt.mode)
			}
		})
	}
}

func TestMouseOptions(t *testing.T) {
	type tc struct {
		opt         AppOption
		wantEnabled bool
	}

	tests := map[string]tc{
		"WithMouse enables mouse explicitly": {
			opt:         WithMouse(),
			wantEnabled: true,
		},
		"WithoutMouse disables mouse explicitly": {
			opt:         WithoutMouse(),
			wantEnabled: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Start from the opposite of the expected value so the assertion
			// proves the option actually flipped the field.
			app := &App{mouseEnabled: !tt.wantEnabled}
			if err := tt.opt(app); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.mouseEnabled != tt.wantEnabled {
				t.Fatalf("mouseEnabled = %v, want %v", app.mouseEnabled, tt.wantEnabled)
			}
			if !app.mouseExplicit {
				t.Fatal("expected mouseExplicit to be true")
			}
		})
	}
}

func TestAppFlagAndCallbackOptions(t *testing.T) {
	globalKeyCalls := 0
	preRenderCalls := 0
	postRenderCalls := 0

	type tc struct {
		opt    AppOption
		assert func(t *testing.T, app *App)
	}

	tests := map[string]tc{
		"WithCursor is a no-op": {
			opt: WithCursor(),
			assert: func(t *testing.T, app *App) {
				if app.manualCursor {
					t.Fatal("WithCursor should not enable manual cursor")
				}
			},
		},
		"WithManualCursor disables framework cursor management": {
			opt: WithManualCursor(),
			assert: func(t *testing.T, app *App) {
				if !app.manualCursor {
					t.Fatal("expected manualCursor to be true")
				}
			},
		},
		"WithLegacyKeyboard forces legacy mode": {
			opt: WithLegacyKeyboard(),
			assert: func(t *testing.T, app *App) {
				if !app.legacyKeyboard {
					t.Fatal("expected legacyKeyboard to be true")
				}
			},
		},
		"WithGlobalKeyHandler stores a working handler": {
			opt: WithGlobalKeyHandler(func(ke KeyEvent) bool {
				globalKeyCalls++
				return ke.Key == KeyEnter
			}),
			assert: func(t *testing.T, app *App) {
				if app.globalKeyHandler == nil {
					t.Fatal("expected globalKeyHandler to be set")
				}
				if !app.globalKeyHandler(KeyEvent{Key: KeyEnter}) {
					t.Fatal("expected handler to consume Enter")
				}
				if app.globalKeyHandler(KeyEvent{Key: KeyEscape}) {
					t.Fatal("expected handler to pass Escape through")
				}
				if globalKeyCalls != 2 {
					t.Fatalf("handler called %d times, want 2", globalKeyCalls)
				}
			},
		},
		"WithPostRenderHook stores a working hook": {
			opt: WithPostRenderHook(func() {
				postRenderCalls++
			}),
			assert: func(t *testing.T, app *App) {
				if app.postRenderHook == nil {
					t.Fatal("expected postRenderHook to be set")
				}
				app.postRenderHook()
				if postRenderCalls != 1 {
					t.Fatalf("hook called %d times, want 1", postRenderCalls)
				}
			},
		},
		"WithPreRenderHook stores a working hook": {
			opt: WithPreRenderHook(func() {
				preRenderCalls++
			}),
			assert: func(t *testing.T, app *App) {
				if app.preRenderHook == nil {
					t.Fatal("expected preRenderHook to be set")
				}
				app.preRenderHook()
				if preRenderCalls != 1 {
					t.Fatalf("hook called %d times, want 1", preRenderCalls)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{}
			if err := tt.opt(app); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.assert(t, app)
		})
	}
}

// optionsTestComponent is a minimal Component used to exercise WithRootComponent.
type optionsTestComponent struct {
	el *Element
}

func (c *optionsTestComponent) Render(app *App) *Element {
	return c.el
}

func TestRootOptions(t *testing.T) {
	rootEl := New()
	viewEl := New()
	view := newMockViewable(viewEl)
	compEl := New()
	comp := &optionsTestComponent{el: compEl}

	type tc struct {
		opt      AppOption
		wantRoot *Element
		wantComp Component
	}

	tests := map[string]tc{
		"WithRoot applies element root": {
			opt:      WithRoot(rootEl),
			wantRoot: rootEl,
		},
		"WithRootView applies viewable root": {
			opt:      WithRootView(view),
			wantRoot: viewEl,
		},
		"WithRootComponent applies component root": {
			opt:      WithRootComponent(comp),
			wantRoot: compEl,
			wantComp: comp,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := newTestApp(80, 24)
			if err := tt.opt(app); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.pendingRootApply == nil {
				t.Fatal("expected pendingRootApply to be set")
			}
			if app.root != nil {
				t.Fatal("root should not be applied before pendingRootApply runs")
			}

			// NewApp runs the pending apply after initialization; simulate that.
			app.pendingRootApply(app)

			if app.root != tt.wantRoot {
				t.Fatalf("root = %p, want %p", app.root, tt.wantRoot)
			}
			if tt.wantComp != nil && app.rootComponent != tt.wantComp {
				t.Fatal("expected rootComponent to be set to the component")
			}
		})
	}
}
