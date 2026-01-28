package main

import "fmt"

func CounterUI(count int) Element {
	<div class="flex-col gap-1 p-2">
		<div class="border-rounded p-1 flex-col items-center justify-center">
			<div class="bg-red w-full h-1">
				<span class="font-bold text-cyan text-center w-full">Counter Examples</span>
			</div>
			<hr class="border border-red" />
			<span>{"Count:"}</span>
			<span class="font-bold text-blue">{fmt.Sprintf("%d", count)}</span>
		</div>
		<br />
		<div class="flex gap-1 justify-center">
			<span class="font-dim">{"Press +/- to change, q to quit"}</span>
		</div>
	</div>
}
