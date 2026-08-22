// Package ui presents the game model as a terminal table: the columns a die is
// shown in, how they sort, the Catppuccin Mocha theme they are drawn in, and
// the Bubble Tea program that puts it all on screen.
//
// # Layout
//
//   - table.go - Column, the column sets each view uses, sorting and search.
//   - theme.go - the palette, the styles built from it, and the block wordmark.
//   - model.go - the Bubble Tea model: state, key handling and rendering.
//
// [Run] is the whole entry point. A command needs nothing from this package but
// a slice of dice to hand it.
//
// # Modes
//
// Tab cycles the three views - raw weights, literal odds, effective odds - and
// m turns multi-select on top of whichever is live. Picking several faces adds
// a total column of the chance of rolling any one of them, counted the way the
// live view counts, and the table is hard sorted on it.
//
// A Column pairs a header with a cell renderer, a comparator and a tint, so the
// value on screen, the order it sorts in and the color it is drawn in cannot
// drift apart as the table grows columns.
//
// This package depends on internal/pkg/game and never the other way round.
package ui
