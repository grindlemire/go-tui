package testdata

import tui "github.com/grindlemire/go-tui"

type myModal struct {
	app        *tui.App
	showModal  *tui.State[bool]
	gameOver   *tui.State[bool]
	confirmBtn tui.Ref
}

func MyModal() *myModal {
	return &myModal{
		showModal: tui.NewState(false),
		gameOver:  tui.NewState(false),
	}
}

func (c *myModal) gameOverKeys() tui.KeyMap {
	return tui.KeyMap{
		tui.OnPreemptStop(tui.Rune('n'), func(ke tui.KeyEvent) {}),
		tui.OnPreemptStop(tui.Rune('q'), func(ke tui.KeyEvent) {}),
	}
}

templ (c *myModal) Render() {
	<div class="flex-col">
		<span>Background content</span>
		<modal open={c.showModal} class="justify-center items-center" backdrop="dim">
			<div class="w-40 border-rounded p-2 flex-col gap-1">
				<span class="font-bold">Are you sure?</span>
				<button ref={c.confirmBtn}>OK</button>
			</div>
		</modal>
		<modal open={c.gameOver} keyMap={c.gameOverKeys()} trapFocus={false}>
			<div class="border-rounded p-2">
				<span>Game Over</span>
			</div>
		</modal>
	</div>
}
