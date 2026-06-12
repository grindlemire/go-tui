package tui

// mountKeyNode chains an outer key with one loop level's key value.
// Keys built from distinct (site, parts...) inputs never compare equal,
// and a node never equals a plain int site key, so call sites cannot
// collide with each other by construction.
type mountKeyNode struct {
	parent any
	part   any
}

// MountKey builds a comparable cache key for a mounted component from its
// generated call-site id and the key values of the enclosing loops,
// outermost first. A key={...} attribute in .gsx replaces the loop values
// with the user's expression. Called by generated code. Every part must be
// a comparable value (slice indices, map keys, or user-provided ids).
func MountKey(site int, parts ...any) any {
	var key any = site
	for _, part := range parts {
		key = mountKeyNode{parent: key, part: part}
	}
	return key
}
