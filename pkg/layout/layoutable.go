package layout

// Layoutable is the interface for anything that can participate in layout calculation.
// The layout engine works entirely with this interface, enabling custom implementations.
type Layoutable interface {
	// LayoutStyle returns the layout style properties for this element.
	LayoutStyle() Style

	// LayoutChildren returns the children to be laid out.
	LayoutChildren() []Layoutable

	// SetLayout is called by the layout engine to store computed layout.
	SetLayout(Layout)

	// GetLayout returns the last computed layout.
	GetLayout() Layout

	// IsDirty returns whether this element needs layout recalculation.
	IsDirty() bool

	// SetDirty marks this element as needing recalculation.
	SetDirty(dirty bool)
}
