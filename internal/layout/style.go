package layout

// Direction specifies the main axis for laying out children.
type Direction uint8

const (
	Row    Direction = iota // Children laid out left-to-right
	Column                  // Children laid out top-to-bottom
)

// Justify specifies how children are distributed along the main axis.
type Justify uint8

const (
	JustifyStart        Justify = iota // Pack at start
	JustifyEnd                         // Pack at end
	JustifyCenter                      // Center children
	JustifySpaceBetween                // Even space between, none at edges
	JustifySpaceAround                 // Even space around each child
	JustifySpaceEvenly                 // Equal space between and at edges
)

// Align specifies how children are positioned on the cross axis.
type Align uint8

const (
	AlignStart   Align = iota // Align to start of cross axis
	AlignEnd                  // Align to end of cross axis
	AlignCenter               // Center on cross axis
	AlignStretch              // Stretch to fill cross axis
)

// Style contains all layout properties for a node.
type Style struct {
	// Sizing
	Width     Value
	Height    Value
	MinWidth  Value
	MinHeight Value
	MaxWidth  Value
	MaxHeight Value

	// Flex container properties
	Direction      Direction
	JustifyContent Justify
	AlignItems     Align
	Gap            int // Space between children (main axis only)

	// Flex item properties
	FlexGrow   float64 // How much to grow relative to siblings
	FlexShrink float64 // How much to shrink relative to siblings (default 1)
	AlignSelf  *Align  // Override parent's AlignItems (nil = inherit)

	// Spacing
	Padding Edges
	Margin  Edges
}

// DefaultStyle returns a Style with sensible defaults.
func DefaultStyle() Style {
	return Style{
		Width:      Auto(),
		Height:     Auto(),
		MinWidth:   Fixed(0),
		MinHeight:  Fixed(0),
		MaxWidth:   Auto(), // No maximum
		MaxHeight:  Auto(), // No maximum
		Direction:  Row,
		AlignItems: AlignStretch,
		FlexShrink: 1.0,
	}
}
