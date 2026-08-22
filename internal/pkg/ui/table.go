package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

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
	// tint colors d's cell, given whether its row is the selected one. It is
	// unexported because a Column is only ever built by the constructors here,
	// which is what guarantees every column has one.
	tint func(d game.Die, selected bool) lipgloss.Style
}

// style is how d's cell in this column is drawn.
func (c Column) style(d game.Die, selected bool) lipgloss.Style {
	if c.tint == nil {
		return pageStyle.Background(rowBG(selected))
	}
	return c.tint(d, selected)
}

// NameColumn is the die's display name.
func NameColumn() Column {
	return Column{
		Title: "Die",
		Width: 24,
		Value: func(d game.Die) string { return d.Name },
		Less:  func(a, b game.Die) bool { return foldName(a.Name) < foldName(b.Name) },
		tint:  func(_ game.Die, selected bool) lipgloss.Style { return nameStyle(selected) },
	}
}

// FaceColumn is the chance of rolling f, counted according to mode.
func FaceColumn(f game.Face, mode game.ProbMode) Column {
	return Column{
		Title: f.String(),
		Width: 7,
		Value: func(d game.Die) string { return FormatPct(d.ProbBy(f, mode)) },
		Less:  func(a, b game.Die) bool { return a.ProbBy(f, mode) < b.ProbBy(f, mode) },
		tint: func(d game.Die, selected bool) lipgloss.Style {
			return probStyle(d.ProbBy(f, mode), selected)
		},
	}
}

// WeightColumn is the raw weight the die gives f, as the game data stores it.
//
// A weight only means anything next to the die's own total, which is why
// [TotalColumn] travels with these.
func WeightColumn(f game.Face) Column {
	return Column{
		Title: f.String(),
		// Weights are small integers, so these are narrower than a percentage
		// column. That keeps the weights view, which carries an extra total
		// column, inside an 80-column terminal.
		Width: 5,
		Value: func(d game.Die) string { return strconv.Itoa(FaceWeight(d, f)) },
		Less:  func(a, b game.Die) bool { return FaceWeight(a, f) < FaceWeight(b, f) },
		// Tinted by the face's share of the die rather than by the weight
		// itself, so a die's bias reads the same in every view even though the
		// numbers on screen change.
		tint: func(d game.Die, selected bool) lipgloss.Style {
			return probStyle(d.Prob(f), selected)
		},
	}
}

// TotalColumn is the sum of a die's side weights, the denominator the other
// weight columns are read against.
func TotalColumn() Column {
	return Column{
		Title: "Total",
		Width: 7,
		Value: func(d game.Die) string { return strconv.Itoa(d.TotalWeight()) },
		Less:  func(a, b game.Die) bool { return a.TotalWeight() < b.TotalWeight() },
		tint:  func(_ game.Die, selected bool) lipgloss.Style { return totalStyle(selected) },
	}
}

// FaceWeight is the weight d puts on f, summed over every side showing it. The
// game data is free to spread one face across several sides, so this is not the
// same as finding the first matching side.
func FaceWeight(d game.Die, f game.Face) int {
	n := 0
	for _, s := range d.Sides {
		if s.Face == f {
			n += s.Weight
		}
	}
	return n
}

// WeightColumns is the name column, one weight column per face, then the total.
func WeightColumns() []Column {
	cols := make([]Column, 0, len(game.Pips)+3)
	cols = append(cols, NameColumn())
	for _, f := range game.Pips {
		cols = append(cols, WeightColumn(f))
	}
	return append(cols, WeightColumn(game.Joker), TotalColumn())
}

// DefaultColumns is the name column and one column per pip, with a joker
// column under [game.Literal].
//
// Under Literal the joker needs a column of its own even though few dice carry
// one: without it that probability mass would be stranded in no column at all,
// and the game is free to add joker dice that put it on a slot other than the 1
// face. Under [game.Effective] the opposite holds - a joker is already counted
// into every pip, so a column of its own would report the same mass twice.
func DefaultColumns(mode game.ProbMode) []Column {
	cols := make([]Column, 0, len(game.Pips)+2)
	cols = append(cols, NameColumn())
	for _, f := range game.Pips {
		cols = append(cols, FaceColumn(f, mode))
	}
	if mode == game.Effective {
		return cols
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
