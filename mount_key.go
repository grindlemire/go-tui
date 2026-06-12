package tui

import (
	"fmt"
	"reflect"
)

// mountKeyNode chains an outer key with one loop level's key value. Distinct
// (site, parts...) inputs never compare equal, so call sites cannot collide.
type mountKeyNode struct {
	parent any
	part   any
}

// MountKey builds a comparable cache key for a mounted component from its
// generated call-site id and the enclosing loops' key values, outermost
// first. Called by generated code. Parts must be comparable values (ints,
// strings, comparable structs, pointers); MountKey panics on a
// non-comparable part such as a slice or map.
func MountKey(site int, parts ...any) any {
	var key any = site
	for _, part := range parts {
		if part != nil && !reflect.TypeOf(part).Comparable() {
			panic(fmt.Sprintf("tui.MountKey: key part of type %T is not comparable; use an int, string, pointer, or comparable struct", part))
		}
		key = mountKeyNode{parent: key, part: part}
	}
	return key
}
