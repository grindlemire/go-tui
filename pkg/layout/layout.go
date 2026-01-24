package layout

// Layout holds the computed position and size after layout calculation.
type Layout struct {
	// Rect is the border box—the space allocated by the parent after
	// applying this node's margin. Use for hit testing and bounds.
	Rect Rect

	// ContentRect is Rect minus padding—the area where children are placed.
	// Use for rendering content and positioning children.
	ContentRect Rect

	// AbsoluteX and AbsoluteY store the true float position before rounding.
	// This enables jitter-free animation by allowing children to compute
	// their absolute position from the parent's float position, then round
	// only once at the final stage (Yoga-style rounding).
	AbsoluteX float64
	AbsoluteY float64
}
