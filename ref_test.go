package tui

import (
	"testing"
)

func TestRef_SetAndGet(t *testing.T) {
	type tc struct {
		setup    func() (*Ref, *Element)
		wantNil  bool
		wantSame bool
	}

	tests := map[string]tc{
		"set and get returns same element": {
			setup: func() (*Ref, *Element) {
				r := NewRef()
				el := New()
				r.Set(el)
				return r, el
			},
			wantNil:  false,
			wantSame: true,
		},
		"overwrite returns latest element": {
			setup: func() (*Ref, *Element) {
				r := NewRef()
				r.Set(New()) // first
				el := New()  // second
				r.Set(el)
				return r, el
			},
			wantNil:  false,
			wantSame: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r, el := tt.setup()
			got := r.El()
			if tt.wantNil && got != nil {
				t.Error("El() should return nil")
			}
			if tt.wantSame && got != el {
				t.Error("El() should return the same element that was Set()")
			}
		})
	}
}

func TestRef_IsSet(t *testing.T) {
	type tc struct {
		setup func() *Ref
		want  bool
	}

	tests := map[string]tc{
		"false before set": {
			setup: func() *Ref {
				return NewRef()
			},
			want: false,
		},
		"true after set": {
			setup: func() *Ref {
				r := NewRef()
				r.Set(New())
				return r
			},
			want: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := tt.setup()
			if r.IsSet() != tt.want {
				t.Errorf("IsSet() = %v, want %v", r.IsSet(), tt.want)
			}
		})
	}
}

func TestRef_NilBeforeSet(t *testing.T) {
	r := NewRef()
	if r.El() != nil {
		t.Error("El() should return nil before Set() is called")
	}
}

func TestRefList_AppendAndAll(t *testing.T) {
	type tc struct {
		count int
	}

	tests := map[string]tc{
		"empty list":       {count: 0},
		"single element":   {count: 1},
		"multiple elements": {count: 5},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := NewRefList()
			elems := make([]*Element, tt.count)
			for i := 0; i < tt.count; i++ {
				elems[i] = New()
				r.Append(elems[i])
			}

			all := r.All()
			if len(all) != tt.count {
				t.Errorf("All() len = %d, want %d", len(all), tt.count)
			}
			for i, el := range all {
				if el != elems[i] {
					t.Errorf("All()[%d] is not the expected element", i)
				}
			}
		})
	}
}

func TestRefList_At(t *testing.T) {
	type tc struct {
		index   int
		wantNil bool
	}

	r := NewRefList()
	el0 := New()
	el1 := New()
	el2 := New()
	r.Append(el0)
	r.Append(el1)
	r.Append(el2)

	tests := map[string]tc{
		"valid index 0":        {index: 0, wantNil: false},
		"valid index 1":        {index: 1, wantNil: false},
		"valid index 2":        {index: 2, wantNil: false},
		"out of bounds high":   {index: 3, wantNil: true},
		"out of bounds negative": {index: -1, wantNil: true},
	}

	elems := []*Element{el0, el1, el2}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := r.At(tt.index)
			if tt.wantNil {
				if got != nil {
					t.Error("At() should return nil for out of bounds index")
				}
			} else {
				if got != elems[tt.index] {
					t.Error("At() should return the correct element")
				}
			}
		})
	}
}

func TestRefList_Len(t *testing.T) {
	type tc struct {
		count int
	}

	tests := map[string]tc{
		"empty":    {count: 0},
		"one":      {count: 1},
		"multiple": {count: 3},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := NewRefList()
			for i := 0; i < tt.count; i++ {
				r.Append(New())
			}
			if r.Len() != tt.count {
				t.Errorf("Len() = %d, want %d", r.Len(), tt.count)
			}
		})
	}
}

func TestRefMap_PutAndGet(t *testing.T) {
	type tc struct {
		key     string
		wantNil bool
	}

	r := NewRefMap[string]()
	elA := New()
	elB := New()
	r.Put("a", elA)
	r.Put("b", elB)

	tests := map[string]tc{
		"existing key a":  {key: "a", wantNil: false},
		"existing key b":  {key: "b", wantNil: false},
		"missing key":     {key: "c", wantNil: true},
	}

	expected := map[string]*Element{"a": elA, "b": elB}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := r.Get(tt.key)
			if tt.wantNil {
				if got != nil {
					t.Error("Get() should return nil for missing key")
				}
			} else {
				if got != expected[tt.key] {
					t.Error("Get() should return the correct element")
				}
			}
		})
	}
}

func TestRefMap_All(t *testing.T) {
	r := NewRefMap[string]()
	elA := New()
	elB := New()
	r.Put("a", elA)
	r.Put("b", elB)

	all := r.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d, want 2", len(all))
	}
	if all["a"] != elA {
		t.Error("All()[\"a\"] should be elA")
	}
	if all["b"] != elB {
		t.Error("All()[\"b\"] should be elB")
	}
}

func TestRefMap_GetMissing(t *testing.T) {
	r := NewRefMap[string]()
	if r.Get("nonexistent") != nil {
		t.Error("Get() should return nil for missing key")
	}
}

func TestRefMap_Len(t *testing.T) {
	type tc struct {
		count int
	}

	tests := map[string]tc{
		"empty":    {count: 0},
		"one":      {count: 1},
		"multiple": {count: 3},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := NewRefMap[int]()
			for i := 0; i < tt.count; i++ {
				r.Put(i, New())
			}
			if r.Len() != tt.count {
				t.Errorf("Len() = %d, want %d", r.Len(), tt.count)
			}
		})
	}
}
