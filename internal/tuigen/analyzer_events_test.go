package tuigen

import "testing"

func TestAnalyzer_DetectEventsVars_GoCodeDeclaration(t *testing.T) {
	input := `package x
templ Example() {
	bus := tui.NewEvents[string]("app.notifications")
	<span>ok</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	eventsVars := analyzer.DetectEventsVars(file.Components[0])
	if len(eventsVars) != 1 {
		t.Fatalf("expected 1 events var, got %d", len(eventsVars))
	}
	if eventsVars[0].Name != "bus" {
		t.Fatalf("expected var name bus, got %s", eventsVars[0].Name)
	}
}

func TestAnalyzer_DetectEventsVars_MultipleDeclarations(t *testing.T) {
	input := `package x
templ Example() {
	eventsA := tui.NewEvents[string]("topic.a")
	eventsB := tui.NewEvents[int]("topic.b")
	<span>ok</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	eventsVars := analyzer.DetectEventsVars(file.Components[0])
	if len(eventsVars) != 2 {
		t.Fatalf("expected 2 events vars, got %d", len(eventsVars))
	}
	if eventsVars[0].Name != "eventsA" || eventsVars[1].Name != "eventsB" {
		t.Fatalf("unexpected events vars: %#v", eventsVars)
	}
}
