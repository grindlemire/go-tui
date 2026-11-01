package tui

// instanceWatchers returns the watchers a component provides. Elements start
// theirs through the root tree walk instead.
func instanceWatchers(comp Component) []Watcher {
	switch c := comp.(type) {
	case *Element:
		return nil
	case WatcherProvider:
		return c.Watchers()
	case Viewable:
		// Generated view structs mounted from a struct component.
		return c.GetWatchers()
	}
	return nil
}

// collectComponentWatchers returns the root component's watchers. Mounted
// components start their watchers when they mount and stop them when they
// unmount (see attach and evict), so only the root needs collecting.
func collectComponentWatchers(rootComp Component) []Watcher {
	return instanceWatchers(rootComp)
}
