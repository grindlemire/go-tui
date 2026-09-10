package tui

// collectComponentWatchers walks the element tree and collects watchers
// from all components that implement WatcherProvider or Viewable.
func collectComponentWatchers(rootComp Component, root *Element) []Watcher {
	var watchers []Watcher

	walkComponents(rootComp, root, func(comp Component) {
		switch c := comp.(type) {
		case WatcherProvider:
			watchers = append(watchers, c.Watchers()...)
		case Viewable:
			// Generated view structs mounted from a struct component.
			watchers = append(watchers, c.GetWatchers()...)
		}
	})

	return watchers
}
