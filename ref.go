package tui

import "sync"

// Ref is a reference to an Element, set during construction
// and accessed later in handlers. Thread-safe.
type Ref struct {
	mu    sync.RWMutex
	value *Element
}

// NewRef creates a new empty Ref.
func NewRef() *Ref {
	return &Ref{}
}

// Set stores the element in this ref.
func (r *Ref) Set(v *Element) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = v
}

// El returns the referenced element, or nil if not yet set.
func (r *Ref) El() *Element {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// IsSet returns true if the ref has been set to a non-nil element.
func (r *Ref) IsSet() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value != nil
}

// RefList holds references to multiple elements created in a loop.
// Thread-safe.
type RefList struct {
	mu    sync.RWMutex
	elems []*Element
}

// NewRefList creates a new empty RefList.
func NewRefList() *RefList {
	return &RefList{}
}

// Append adds an element to this ref list.
func (r *RefList) Append(el *Element) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.elems = append(r.elems, el)
}

// All returns a copy of all referenced elements.
func (r *RefList) All() []*Element {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Element, len(r.elems))
	copy(out, r.elems)
	return out
}

// At returns the element at the given index, or nil if out of bounds.
func (r *RefList) At(i int) *Element {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if i < 0 || i >= len(r.elems) {
		return nil
	}
	return r.elems[i]
}

// Len returns the number of elements in this ref list.
func (r *RefList) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.elems)
}

// RefMap holds keyed references to elements created in a loop.
// Thread-safe.
type RefMap[K comparable] struct {
	mu    sync.RWMutex
	elems map[K]*Element
}

// NewRefMap creates a new empty RefMap.
func NewRefMap[K comparable]() *RefMap[K] {
	return &RefMap[K]{elems: make(map[K]*Element)}
}

// Put stores an element with the given key.
func (r *RefMap[K]) Put(key K, el *Element) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.elems[key] = el
}

// Get returns the element for the given key, or nil if not found.
func (r *RefMap[K]) Get(key K) *Element {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.elems[key]
}

// All returns a copy of all keyed elements.
func (r *RefMap[K]) All() map[K]*Element {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[K]*Element, len(r.elems))
	for k, v := range r.elems {
		out[k] = v
	}
	return out
}

// Len returns the number of elements in this ref map.
func (r *RefMap[K]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.elems)
}
