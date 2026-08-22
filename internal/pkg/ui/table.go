// Package ui turns the game model into something a terminal table can show:
// columns, sorting and search. It depends on internal/pkg/game and never the
// other way round.
//
// Nothing here is bound to a particular TUI framework yet. A Column is a header,
// a cell renderer and a comparator; whatever draws it is free to decide how.
package ui

import (
	"fmt"
	"slices"
	"strings"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// Column is the single source of truth for one table column: its header, how a
// die renders in it, and how two dice compare in it.
//
// Keeping rendering and ordering together stops the displayed value and the
// sort key from drifting apart, which is the usual failure mode when a table
// grows columns.
type Column struct {
	Title string
	Width int
	// Value renders d's cell.
	Value func(d game.Die) string
	// Less orders ascending. SortBy inverts it for descending order.
	Less func(a, b game.Die) bool
}

// NameColumn is the die's display name.
func NameColumn() Column {
	return Column{
		Title: "Die",
		Width: 24,
		Value: func(d game.Die) string { return d.Name },
		Less:  func(a, b game.Die) bool { return foldName(a.Name) < foldName(b.Name) },
	}
}

// FaceColumn is the chance of rolling f, counted according to mode.
func FaceColumn(f game.Face, mode game.ProbMode) Column {
	return Column{
		Title: f.String(),
		Width: 7,
		Value: func(d game.Die) string { return FormatPct(d.ProbBy(f, mode)) },
		Less:  func(a, b game.Die) bool { return a.ProbBy(f, mode) < b.ProbBy(f, mode) },
	}
}

// DefaultColumns is the name column, one column per pip, then the joker column.
//
// The joker column is always present, even though only a few dice currently
// carry one. In game.Literal mode its absence would strand that probability
// mass in no column at all, and the game is free to add joker dice that put it
// on a slot other than the 1 face.
func DefaultColumns(mode game.ProbMode) []Column {
	cols := make([]Column, 0, len(game.Pips)+2)
	cols = append(cols, NameColumn())
	for _, f := range game.Pips {
		cols = append(cols, FaceColumn(f, mode))
	}
	return append(cols, FaceColumn(game.Joker, mode))
}

// SortBy orders dice in place by col. The sort is stable, so the previous
// ordering survives as a tiebreak between equal values.
func SortBy(dice []game.Die, col Column, desc bool) {
	slices.SortStableFunc(dice, func(a, b game.Die) int {
		switch {
		case col.Less(a, b):
			if desc {
				return 1
			}
			return -1
		case col.Less(b, a):
			if desc {
				return -1
			}
			return 1
		}
		return 0
	})
}

// Filter returns the dice whose name contains query, case-insensitively. An
// empty query returns everything.
func Filter(dice []game.Die, query string) []game.Die {
	q := foldName(strings.TrimSpace(query))
	if q == "" {
		return slices.Clone(dice)
	}
	out := make([]game.Die, 0, len(dice))
	for _, d := range dice {
		if strings.Contains(foldName(d.Name), q) {
			out = append(out, d)
		}
	}
	return out
}

// FormatPct renders a probability as a percentage.
func FormatPct(p float64) string { return fmt.Sprintf("%.1f%%", p*100) }

// nameFolder normalises the typographic apostrophes the game data uses, so
// that searching "tengri's" matches "Tengri’s die".
var nameFolder = strings.NewReplacer("’", "'", "‘", "'", "`", "'")

func foldName(s string) string { return nameFolder.Replace(strings.ToLower(s)) }
