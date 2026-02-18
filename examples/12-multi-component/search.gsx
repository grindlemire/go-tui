package main

import (
	"fmt"
	"strings"
	tui "github.com/grindlemire/go-tui"
)

type searchBar struct {
	active *tui.State[bool]
	query  *tui.State[string]
}

func SearchBar(active *tui.State[bool], query *tui.State[string]) *searchBar {
	return &searchBar{active: active, query: query}
}

func (s *searchBar) KeyMap() tui.KeyMap {
	if !s.active.Get() {
		return nil
	}
	return tui.KeyMap{
		tui.OnRunesStop(s.appendChar),
		tui.OnKeyStop(tui.KeyBackspace, s.deleteChar),
		tui.OnKeyStop(tui.KeyEnter, s.submit),
		tui.OnKeyStop(tui.KeyEscape, s.deactivate),
	}
}

func (s *searchBar) appendChar(ke tui.KeyEvent) {
	s.query.Set(s.query.Get() + string(ke.Rune))
}

func (s *searchBar) deleteChar(ke tui.KeyEvent) {
	q := s.query.Get()
	if len(q) > 0 {
		s.query.Set(q[:len(q)-1])
	}
}

func (s *searchBar) submit(ke tui.KeyEvent) {
	s.active.Set(false)
}

func (s *searchBar) deactivate(ke tui.KeyEvent) {
	s.active.Set(false)
	s.query.Set("")
}

templ (s *searchBar) Render() {
	<div class="shrink-0">
		@if s.active.Get() {
			<div class="border-rounded border-cyan p-1 flex gap-1">
				<span class="text-cyan font-bold">Search:</span>
				<span>{s.query.Get()}</span>
				<span class="text-cyan font-bold">|</span>
			</div>
		}
	</div>
}

// Content displays files for the selected category
type content struct {
	category    *tui.State[string]
	categoryBus *tui.Events[string]
	query       *tui.State[string]
}

func Content(query *tui.State[string]) *content {
	c := &content{
		category:    tui.NewState("Documents"),
		categoryBus: tui.NewEvents[string](categoryTopic),
		query:       query,
	}
	c.categoryBus.Subscribe(c.onCategoryChanged)
	return c
}

func (c *content) onCategoryChanged(category string) {
	c.category.Set(category)
}

var filesByCategory = map[string][]string{
	"Documents": {"report.pdf", "notes.md", "budget.xlsx", "readme.txt", "design.doc"},
	"Images":    {"photo.jpg", "screenshot.png", "logo.svg", "banner.gif", "icon.ico"},
	"Music":     {"song.mp3", "album.flac", "podcast.ogg", "ringtone.wav"},
	"Projects":  {"go-tui/", "website/", "api-server/", "mobile-app/", "scripts/"},
	"Downloads": {"setup.exe", "archive.zip", "data.csv", "patch-v2.tar.gz"},
}

func (c *content) filteredFiles() []string {
	cat := c.category.Get()
	files, ok := filesByCategory[cat]
	if !ok {
		return nil
	}
	q := strings.ToLower(c.query.Get())
	if q == "" {
		return files
	}
	var result []string
	for _, f := range files {
		if strings.Contains(strings.ToLower(f), q) {
			result = append(result, f)
		}
	}
	return result
}

templ (c *content) Render() {
	<div class="flex-col flex-grow p-1 gap-1">
		<span class="font-bold text-cyan">{c.category.Get() + "/"}</span>
		<hr />
		@for i, file := range c.filteredFiles() {
			@if i == len(c.filteredFiles())-1 {
				<span>{fmt.Sprintf("  └── %s", file)}</span>
			} @else {
				<span>{fmt.Sprintf("  ├── %s", file)}</span>
			}
		}
		@if len(c.filteredFiles()) == 0 {
			<span class="font-dim">No matching files</span>
		}
		@if c.query.Get() != "" {
			<br />
			<span class="font-dim">{fmt.Sprintf("Filtering: \"%s\"", c.query.Get())}</span>
		}
	</div>
}
