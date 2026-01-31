// Package layout implements a pure-Go flexbox layout engine for terminal UIs.
//
// It supports row/column directions, justify and align modes, padding, margin,
// gap, min/max constraints, percentage and fixed dimensions, and intrinsic sizing.
// Types are re-exported through the root tui package for public consumption.
//
// The main entry point is [Calculate], which takes a [Layoutable] tree and
// computes absolute [Rect] positions for each node.
package layout
