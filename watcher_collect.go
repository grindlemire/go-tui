package tui

// collectComponentWatchers walks the element tree and collects watchers
// from all components that implement WatcherProvider.
func collectComponentWatchers(rootComp Component, root *Element) []Watcher {
	var watchers []Watcher

	walkComponents(rootComp, root, func(comp Component) {
		if wp, ok := comp.(WatcherProvider); ok {
			watchers = append(watchers, wp.Watchers()...)
		}
	})

	return watchers
}
