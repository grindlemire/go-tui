package layout

// Unit specifies how a Value is interpreted.
type Unit uint8

const (
	UnitAuto    Unit = iota // Size determined by content/flex
	UnitFixed               // Absolute terminal cells
	UnitPercent             // Percentage of parent's available space
)

// Value represents a dimension that can be fixed, percentage, or auto.
type Value struct {
	Amount float64
	Unit   Unit
}

// Auto returns a Value that should be computed from content/flex.
func Auto() Value {
	return Value{Unit: UnitAuto}
}

// Fixed returns a Value representing an absolute number of terminal cells.
func Fixed(n int) Value {
	return Value{Amount: float64(n), Unit: UnitFixed}
}

// Percent returns a Value representing a percentage of available space.
// The value is on a 0-100 scale (50.0 = 50%).
func Percent(p float64) Value {
	return Value{Amount: p, Unit: UnitPercent}
}

// Resolve computes the actual integer value given available space.
// For UnitAuto, returns the fallback value.
func (v Value) Resolve(available, fallback int) int {
	switch v.Unit {
	case UnitFixed:
		return int(v.Amount)
	case UnitPercent:
		return int(float64(available) * v.Amount / 100.0)
	case UnitAuto:
		return fallback
	default:
		return fallback
	}
}

// IsAuto returns true if this value should be computed from content/flex.
func (v Value) IsAuto() bool {
	return v.Unit == UnitAuto
}
