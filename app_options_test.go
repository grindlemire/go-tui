package tui

import "testing"

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
