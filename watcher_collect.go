package tui

// collectComponentWatchers walks the element tree and collects watchers
// from all components that implement WatcherProvider.
func collectComponentWatchers(root *Element) []Watcher {
	var watchers []Watcher

	walkComponents(root, func(comp Component) {
		if wp, ok := comp.(WatcherProvider); ok {
			watchers = append(watchers, wp.Watchers()...)
		}
	})

	return watchers
}
