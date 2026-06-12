package tui

import "testing"

func TestMountKey_Equality(t *testing.T) {
	type tc struct {
		a, b      any
		wantEqual bool
	}

	tests := map[string]tc{
		"same site and parts are equal": {
			a: MountKey(1, 2), b: MountKey(1, 2), wantEqual: true,
		},
		"same site and string parts are equal": {
			a: MountKey(0, "alpha"), b: MountKey(0, "alpha"), wantEqual: true,
		},
		"different parts differ": {
			a: MountKey(0, "alpha"), b: MountKey(0, "beta"), wantEqual: false,
		},
		"different sites differ": {
			a: MountKey(0, "x"), b: MountKey(1, "x"), wantEqual: false,
		},
		"issue 88 collision is impossible: loop iteration vs standalone site": {
			a: MountKey(0, 1), b: MountKey(1), wantEqual: false,
		},
		"nested loop iteration vs flat sibling loop iteration differ": {
			a: MountKey(0, 1, 0), b: MountKey(1, 0), wantEqual: false,
		},
		"int part and string part differ": {
			a: MountKey(0, 1), b: MountKey(0, "1"), wantEqual: false,
		},
		"no parts returns the plain site key": {
			a: MountKey(3), b: any(3), wantEqual: true,
		},
		"nested key order matters": {
			a: MountKey(0, 1, 2), b: MountKey(0, 2, 1), wantEqual: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a == tt.b; got != tt.wantEqual {
				t.Errorf("(%v == %v) = %v, want %v", tt.a, tt.b, got, tt.wantEqual)
			}
		})
	}
}

func TestMountKey_UsableAsMapKey(t *testing.T) {
	m := map[any]string{}
	m[MountKey(0, "a")] = "first"
	m[MountKey(0, "b")] = "second"
	m[MountKey(1)] = "third"

	if len(m) != 3 {
		t.Fatalf("expected 3 distinct map entries, got %d", len(m))
	}
	if m[MountKey(0, "a")] != "first" {
		t.Errorf("lookup with rebuilt key failed")
	}
}
