// DEPRECATED: This package is a temporary compatibility shim.
// Use "github.com/grindlemire/go-tui" instead.
//
// This file re-exports all types from internal/layout so that existing code
// importing "github.com/grindlemire/go-tui/pkg/layout" continues to compile.
// It will be removed once all consumers are migrated.
package layout

import "github.com/grindlemire/go-tui/internal/layout"

// Types
type Direction = layout.Direction
type Justify = layout.Justify
type Align = layout.Align
type Style = layout.Style
type Unit = layout.Unit
type Value = layout.Value
type Layoutable = layout.Layoutable
type Size = layout.Size
type Rect = layout.Rect
type Edges = layout.Edges
type Point = layout.Point
type Layout = layout.Layout

// Direction constants
const (
	Row    = layout.Row
	Column = layout.Column
)

// Justify constants
const (
	JustifyStart        = layout.JustifyStart
	JustifyEnd          = layout.JustifyEnd
	JustifyCenter       = layout.JustifyCenter
	JustifySpaceBetween = layout.JustifySpaceBetween
	JustifySpaceAround  = layout.JustifySpaceAround
	JustifySpaceEvenly  = layout.JustifySpaceEvenly
)

// Align constants
const (
	AlignStart   = layout.AlignStart
	AlignEnd     = layout.AlignEnd
	AlignCenter  = layout.AlignCenter
	AlignStretch = layout.AlignStretch
)

// Unit constants
const (
	UnitAuto    = layout.UnitAuto
	UnitFixed   = layout.UnitFixed
	UnitPercent = layout.UnitPercent
)

// Constructors

func Auto() Value              { return layout.Auto() }
func Fixed(n int) Value        { return layout.Fixed(n) }
func Percent(p float64) Value  { return layout.Percent(p) }
func DefaultStyle() Style      { return layout.DefaultStyle() }
func NewRect(x, y, w, h int) Rect { return layout.NewRect(x, y, w, h) }
func EdgeAll(n int) Edges         { return layout.EdgeAll(n) }
func EdgeSymmetric(v, h int) Edges { return layout.EdgeSymmetric(v, h) }
func EdgeTRBL(t, r, b, l int) Edges { return layout.EdgeTRBL(t, r, b, l) }

// Calculate performs layout on the tree rooted at root.
func Calculate(root Layoutable, availableWidth, availableHeight int) {
	layout.Calculate(root, availableWidth, availableHeight)
}
