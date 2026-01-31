package layout

import "testing"

func TestValue_Constructors(t *testing.T) {
	type tc struct {
		value  Value
		isAuto bool
		unit   Unit
		amount float64
	}

	tests := map[string]tc{
		"Auto": {
			value:  Auto(),
			isAuto: true,
			unit:   UnitAuto,
			amount: 0,
		},
		"Fixed": {
			value:  Fixed(100),
			isAuto: false,
			unit:   UnitFixed,
			amount: 100,
		},
		"Percent": {
			value:  Percent(50),
			isAuto: false,
			unit:   UnitPercent,
			amount: 50,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.value.IsAuto(); got != tt.isAuto {
				t.Errorf("IsAuto() = %v, want %v", got, tt.isAuto)
			}
			if tt.value.Unit != tt.unit {
				t.Errorf("Unit = %v, want %v", tt.value.Unit, tt.unit)
			}
			if tt.value.Amount != tt.amount {
				t.Errorf("Amount = %v, want %v", tt.value.Amount, tt.amount)
			}
		})
	}
}

func TestValue_Resolve_Fixed(t *testing.T) {
	type tc struct {
		value     Value
		available int
		fallback  int
		expected  int
	}

	tests := map[string]tc{
		"fixed ignores available": {
			value:     Fixed(50),
			available: 100,
			fallback:  0,
			expected:  50,
		},
		"fixed ignores fallback": {
			value:     Fixed(50),
			available: 100,
			fallback:  999,
			expected:  50,
		},
		"fixed zero": {
			value:     Fixed(0),
			available: 100,
			fallback:  50,
			expected:  0,
		},
		"fixed negative": {
			value:     Fixed(-10),
			available: 100,
			fallback:  50,
			expected:  -10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.value.Resolve(tt.available, tt.fallback)
			if got != tt.expected {
				t.Errorf("Resolve(%d, %d) = %d, want %d",
					tt.available, tt.fallback, got, tt.expected)
			}
		})
	}
}

func TestValue_Resolve_Percent(t *testing.T) {
	type tc struct {
		value     Value
		available int
		fallback  int
		expected  int
	}

	tests := map[string]tc{
		"50 percent of 100": {
			value:     Percent(50),
			available: 100,
			fallback:  0,
			expected:  50,
		},
		"100 percent of 80": {
			value:     Percent(100),
			available: 80,
			fallback:  0,
			expected:  80,
		},
		"25 percent of 200": {
			value:     Percent(25),
			available: 200,
			fallback:  0,
			expected:  50,
		},
		"0 percent": {
			value:     Percent(0),
			available: 100,
			fallback:  50,
			expected:  0,
		},
		"percent of zero available": {
			value:     Percent(50),
			available: 0,
			fallback:  50,
			expected:  0,
		},
		"fractional percent rounds down": {
			value:     Percent(33.33),
			available: 100,
			fallback:  0,
			expected:  33,
		},
		"percent over 100": {
			value:     Percent(150),
			available: 100,
			fallback:  0,
			expected:  150,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.value.Resolve(tt.available, tt.fallback)
			if got != tt.expected {
				t.Errorf("Resolve(%d, %d) = %d, want %d",
					tt.available, tt.fallback, got, tt.expected)
			}
		})
	}
}

func TestValue_Resolve_Auto(t *testing.T) {
	type tc struct {
		value     Value
		available int
		fallback  int
		expected  int
	}

	tests := map[string]tc{
		"auto returns fallback": {
			value:     Auto(),
			available: 100,
			fallback:  50,
			expected:  50,
		},
		"auto with zero fallback": {
			value:     Auto(),
			available: 100,
			fallback:  0,
			expected:  0,
		},
		"auto ignores available": {
			value:     Auto(),
			available: 999,
			fallback:  42,
			expected:  42,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.value.Resolve(tt.available, tt.fallback)
			if got != tt.expected {
				t.Errorf("Resolve(%d, %d) = %d, want %d",
					tt.available, tt.fallback, got, tt.expected)
			}
		})
	}
}

func TestValue_IsAuto(t *testing.T) {
	type tc struct {
		value    Value
		expected bool
	}

	tests := map[string]tc{
		"Auto is auto": {
			value:    Auto(),
			expected: true,
		},
		"Fixed is not auto": {
			value:    Fixed(10),
			expected: false,
		},
		"Percent is not auto": {
			value:    Percent(50),
			expected: false,
		},
		"zero value is auto": {
			value:    Value{},
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.value.IsAuto(); got != tt.expected {
				t.Errorf("IsAuto() = %v, want %v", got, tt.expected)
			}
		})
	}
}
