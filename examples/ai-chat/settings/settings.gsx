package settings

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ (s *SettingsApp) Render() {
	<div class="flex-col h-full p-2 gap-2">
		<div class="border-rounded p-1" height={3} direction={tui.Row} justify={tui.JustifyCenter} align={tui.AlignCenter}>
			<span class="text-gradient-cyan-magenta font-bold">{"  Settings"}</span>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(0)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Provider"}</span>
				<div class="flex gap-2">
					@for _, p := range s.state.AvailableProviders {
						@if p == s.state.Provider.Get() {
							<span class="text-cyan font-bold">{"● " + p}</span>
						} @else {
							<span class="font-dim">{"○ " + p}</span>
						}
					}
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(1)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Model"}</span>
				<div class="flex gap-2">
					@for _, m := range s.state.ProviderModels[s.state.Provider.Get()] {
						@if m == s.state.Model.Get() {
							<span class="text-cyan font-bold">{"● " + m}</span>
						} @else {
							<span class="font-dim">{"○ " + m}</span>
						}
					}
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(2)} padding={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"Temperature"}</span>
				<div class="flex gap-2 items-center">
					<span class="text-white">{s.tempBar()}</span>
					<span class="text-cyan">{fmt.Sprintf("%.1f", s.state.Temperature.Get())}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">{"← precise"}</span>
					<span class="font-dim">{"creative →"}</span>
				</div>
			</div>
		</div>

		<div border={tui.BorderRounded} borderStyle={s.borderStyleForSection(3)} padding={1} flexGrow={1}>
			<div class="flex-col gap-1">
				<span class="font-bold text-cyan">{"System Prompt"}</span>
				<span class="text-white">{s.state.SystemPrompt.Get()}</span>
			</div>
		</div>

		<div class="flex justify-center gap-2">
			<button ref={s.saveBtn} class="border-rounded border-cyan p-1">{"  Save  "}</button>
			<button ref={s.cancelBtn} class="border-rounded p-1">{"  Cancel  "}</button>
		</div>

		<div class="flex justify-center">
			<span class="font-dim">{"Tab: navigate  ←/→: select  Enter: save  Esc: cancel"}</span>
		</div>
	</div>
}
