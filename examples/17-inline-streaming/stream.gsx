package main

import (
	"fmt"
	"math/rand/v2"
	"time"
	tui "github.com/grindlemire/go-tui"
)

var gradient = tui.NewGradient(tui.BrightCyan, tui.BrightMagenta)

// phrases to cycle through when the user presses Enter.
var phrases = []string{
	"The quick brown fox jumps over the lazy dog.",
	"Stars scattered across the midnight canvas, each one a whisper of ancient light.",
	"Line one of a multi-line message.\nLine two continues here.\nAnd line three wraps it up.",
	"Streaming text appears character by character, just like a real-time API response.",
}

type streamDemo struct {
	app       *tui.App
	phraseIdx int
	streaming *tui.State[bool]
}

func StreamDemo() *streamDemo {
	return &streamDemo{
		streaming: tui.NewState(false),
	}
}

type person struct {
	Name   string
	Role   string
	Status string
	Score  int
}

var (
	allNames   = []string{"Alice", "Bob", "Carol", "Dave", "Eve", "Frank"}
	allRoles   = []string{"Engineer", "Designer", "PM", "Analyst", "DevOps", "QA"}
	allStatuses = []string{"Active", "Away", "Busy", "Offline"}
	statusColors = map[string]tui.Color{
		"Active":  tui.Green,
		"Away":    tui.Yellow,
		"Busy":    tui.Red,
		"Offline": tui.BrightBlack,
	}
)

func randomPeople() []person {
	n := 3 + rand.IntN(3)
	used := map[int]bool{}
	people := make([]person, 0, n)
	for range n {
		idx := rand.IntN(len(allNames))
		for used[idx] {
			idx = rand.IntN(len(allNames))
		}
		used[idx] = true
		people = append(people, person{
			Name:   allNames[idx],
			Role:   allRoles[rand.IntN(len(allRoles))],
			Status: allStatuses[rand.IntN(len(allStatuses))],
			Score:  50 + rand.IntN(51),
		})
	}
	return people
}

templ ReportCard(people []person) {
	<div class="flex justify-center">
		<div class="flex-col border-rounded w-3/4 px-1" borderStyle={tui.NewStyle().Foreground(tui.BrightCyan)}>
			<span class="font-bold text-bright-magenta">Streaming Report</span>
			<table>
				<tr>
					<th class="grow">Name</th>
					<th class="grow">Role</th>
					<th class="grow">Status</th>
					<th>Score</th>
				</tr>
				<hr />
				@for _, p := range people {
					<tr>
						<td class="text-cyan grow">{p.Name}</td>
						<td class="grow">{p.Role}</td>
						<td class="grow" textStyle={tui.NewStyle().Foreground(statusColors[p.Status])}>{p.Status}</td>
						<td class="font-bold">{fmt.Sprintf("%d", p.Score)}</td>
					</tr>
				}
			</table>
		</div>
	</div>
}

// streamWithElement opens a StreamAbove writer, writes gradient intro text,
// inserts a styled card with a randomized table via WriteElement, then writes
// gradient outro text.
func (s *streamDemo) streamWithElement() {
	if s.streaming.Get() {
		return
	}
	s.streaming.Set(true)

	go func() {
		w := s.app.StreamAbove()
		w.WriteGradient("Here's a summary:\n", gradient)
		time.Sleep(100 * time.Millisecond)

		w.WriteElement(ReportCard(randomPeople()))

		time.Sleep(100 * time.Millisecond)
		w.WriteGradient("Done!\n", gradient)
		w.Close()
		s.app.QueueUpdate(func() {
			s.streaming.Set(false)
		})
	}()
}

// streamPhrase opens a StreamAbove writer, writes each character with a
// gradient color and a small delay, then closes the writer.
func (s *streamDemo) streamPhrase() {
	if s.streaming.Get() {
		return
	}
	s.streaming.Set(true)

	text := phrases[s.phraseIdx%len(phrases)]
	s.phraseIdx++

	go func() {
		w := s.app.StreamAbove()
		for _, r := range text {
			w.WriteGradient(string(r), gradient)
			time.Sleep(30 * time.Millisecond)
		}
		w.Close()
		s.app.QueueUpdate(func() {
			s.streaming.Set(false)
		})
	}()
}

func (s *streamDemo) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKeyStop(tui.KeyEnter, func(ke tui.KeyEvent) { s.streamPhrase() }),
		tui.OnKeyStop(tui.KeyTab, func(ke tui.KeyEvent) { s.streamWithElement() }),
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

func (s *streamDemo) statusText() string {
	if s.streaming.Get() {
		return "streaming..."
	}
	return "Enter to stream  |  Tab to stream with element  |  Esc to quit"
}

templ (s *streamDemo) Render() {
	<div class="border-rounded border-cyan items-center justify-center">
		<span class="text-cyan">{s.statusText()}</span>
	</div>
}
