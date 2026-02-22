package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type fileList struct {
	files    []string
	selected *tui.State[int]
	scrollY  *tui.State[int]
	content  *tui.Ref
}

func FileList() *fileList {
	files := []string{
		"main.go", "app.go", "app_loop.go", "app_render.go",
		"buffer.go", "cell.go", "color.go", "component.go",
		"dirty.go", "element.go", "element_options.go",
		"element_render.go", "element_scroll.go", "escape.go",
		"event.go", "focus.go", "key.go", "keymap.go",
		"layout.go", "mount.go", "ref.go", "render.go",
		"state.go", "style.go", "terminal.go", "watcher.go",
	}
	return &fileList{
		files:    files,
		selected: tui.NewState(0),
		scrollY:  tui.NewState(0),
		content:  tui.NewRef(),
	}
}

func (f *fileList) scrollBy(delta int) {
	el := f.content.El()
	if el == nil {
		return
	}
	_, maxY := el.MaxScroll()
	newY := f.scrollY.Get() + delta
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}
	f.scrollY.Set(newY)
}

func (f *fileList) moveTo(idx int) {
	if idx < 0 {
		idx = len(f.files) - 1
	}
	if idx >= len(f.files) {
		idx = 0
	}
	f.selected.Set(idx)

	// Keep selected item visible by adjusting scroll
	el := f.content.El()
	if el == nil {
		return
	}
	_, vpH := el.ViewportSize()
	y := f.scrollY.Get()
	if idx < y {
		f.scrollY.Set(idx)
	} else if idx >= y+vpH {
		f.scrollY.Set(idx - vpH + 1)
	}
}

func (f *fileList) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 1) }),
		tui.OnRune('k', func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 1) }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 1) }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 1) }),
		tui.OnKey(tui.KeyPageDown, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 10) }),
		tui.OnKey(tui.KeyPageUp, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 10) }),
		tui.OnKey(tui.KeyHome, func(ke tui.KeyEvent) { f.moveTo(0) }),
		tui.OnKey(tui.KeyEnd, func(ke tui.KeyEvent) { f.moveTo(len(f.files) - 1) }),
	}
}

func (f *fileList) HandleMouse(me tui.MouseEvent) bool {
	switch me.Button {
	case tui.MouseWheelUp:
		f.scrollBy(-3)
		return true
	case tui.MouseWheelDown:
		f.scrollBy(3)
		return true
	}
	return false
}

templ (f *fileList) Render() {
	<div class="flex-col p-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Files</span>
		<div class="overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
			height={12} ref={f.content} scrollOffset={0, f.scrollY.Get()}>
			@for i, name := range f.files {
				@if i == f.selected.Get() {
					<span class="text-cyan font-bold bg-bright-black">{fmt.Sprintf(" > %s ", name)}</span>
				} @else {
					<span>{fmt.Sprintf("   %s", name)}</span>
				}
			}
		</div>
		<div class="flex justify-between">
			<span class="font-dim">j/k navigate | esc quit</span>
			<span class="font-dim">{fmt.Sprintf("%d/%d", f.selected.Get()+1, len(f.files))}</span>
		</div>
	</div>
}
