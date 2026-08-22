package ui

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// contrast is the WCAG contrast ratio between two hex colors, used to check
// that a cell is actually readable on the background it lands on rather than
// merely a different color from it.
func contrast(a, b lipgloss.Color) float64 {
	lum := func(c lipgloss.Color) float64 {
		h := strings.TrimPrefix(string(c), "#")
		var out float64
		for i, weight := range []float64{0.2126, 0.7152, 0.0722} {
			v, err := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
			if err != nil {
				return math.NaN()
			}
			f := float64(v) / 255
			if f <= 0.03928 {
				f /= 12.92
			} else {
				f = math.Pow((f+0.055)/1.055, 2.4)
			}
			out += weight * f
		}
		return out
	}
	hi, lo := lum(a), lum(b)
	if hi < lo {
		hi, lo = lo, hi
	}
	return (hi + 0.05) / (lo + 0.05)
}

func TestProbStyleSeparatesTheFourCases(t *testing.T) {
	// A face never rolled, one rolled less than a fair die, one about fair, and
	// one favoured all have to be tellable apart at a glance.
	got := map[string]lipgloss.TerminalColor{
		"zero":    probStyle(0, false).GetForeground(),
		"below":   probStyle(uniform/2, false).GetForeground(),
		"uniform": probStyle(uniform, false).GetForeground(),
		"above":   probStyle(uniform*2, false).GetForeground(),
	}
	seen := map[lipgloss.TerminalColor]string{}
	for name, c := range got {
		if prev, dup := seen[c]; dup {
			t.Errorf("%s and %s are the same color %v", prev, name, c)
		}
		seen[c] = name
	}

	// The band around a fair die is deliberately wide enough to absorb float
	// noise, so a hair either side of it still reads as ordinary.
	assert.Equal(t, got["uniform"], probStyle(uniform+0.005, false).GetForeground())
	assert.Equal(t, got["uniform"], probStyle(uniform-0.005, false).GetForeground())
}

func TestZeroStaysLegibleOnASelectedRow(t *testing.T) {
	// Regression: the zero color and the selection background were adjacent
	// shades of the same ramp, so a zero vanished when its row was selected.
	for _, selected := range []bool{false, true} {
		fg, ok := probStyle(0, selected).GetForeground().(lipgloss.Color)
		require.True(t, ok)
		bg := rowBG(selected)
		assert.Greater(t, contrast(fg, bg), 2.0,
			"a zero on a selected=%v row is unreadable at %.2f:1", selected, contrast(fg, bg))
	}

	// It should recede by about as much either way, rather than flaring up when
	// the cursor lands on it.
	unsel := contrast(probStyle(0, false).GetForeground().(lipgloss.Color), rowBG(false))
	sel := contrast(probStyle(0, true).GetForeground().(lipgloss.Color), rowBG(true))
	assert.InDelta(t, unsel, sel, 0.5)
}

func TestSelectionChangesEveryCellBackground(t *testing.T) {
	// Whatever a cell's own color, a selected row must read as one band.
	dice := parseFixture(t)
	for _, c := range append(DefaultColumns(game.Literal), WeightColumns()...) {
		on := c.style(dice[0], true).GetBackground()
		off := c.style(dice[0], false).GetBackground()
		assert.NotEqual(t, on, off, "column %q does not react to selection", c.Title)
	}
}

func TestHeaderAndRadioMarkTheLiveOne(t *testing.T) {
	assert.NotEqual(t, headerStyle(true).GetBackground(), headerStyle(false).GetBackground())
	assert.NotEqual(t, radioStyle(true).GetForeground(), radioStyle(false).GetForeground())
	// Both use the same accent, so "this is the live one" looks the same
	// wherever it appears.
	assert.Equal(t, headerStyle(true).GetBackground(), radioStyle(true).GetForeground())
}

func TestWordmarkRowsStayAligned(t *testing.T) {
	m := wordmark("FARKLE DICE")
	require.NotEmpty(t, m[0])
	// The gradient steps by character position, so the rows must be the same
	// length or the colors will not line up vertically.
	assert.Equal(t, len([]rune(m[0])), len([]rune(m[1])))
	assert.Equal(t, len([]rune(m[1])), len([]rune(m[2])))

	// Every glyph is square too, which is what keeps that true for any word.
	for r, g := range blockFont {
		assert.Equal(t, len([]rune(g[0])), len([]rune(g[1])), "glyph %q", r)
		assert.Equal(t, len([]rune(g[1])), len([]rune(g[2])), "glyph %q", r)
	}
}

func TestWordmarkSkipsUnknownLetters(t *testing.T) {
	// A letter with no glyph is dropped rather than punching a ragged hole in
	// one row only.
	with := wordmark("FA")
	without := wordmark("FZA")
	assert.Equal(t, with, without)
	assert.Equal(t, [3]string{"", "", ""}, wordmark("ZZZ"))
}

func TestRampAtSweepsAndWraps(t *testing.T) {
	const n = 24
	at := func(phase int) []lipgloss.Color {
		out := make([]lipgloss.Color, n)
		for i := range out {
			out[i] = rampAt(i, n, phase)
		}
		return out
	}

	// Raising the phase by one shifts the whole sweep one place left.
	base, next := at(0), at(1)
	assert.Equal(t, base[1:], next[:n-1])

	// And a full lap comes back to where it started, so the loop has no seam.
	assert.Equal(t, base, at(n))

	// Every color drawn is one from the palette.
	for _, c := range base {
		assert.Contains(t, ramp, c)
	}
	assert.NotPanics(t, func() { rampAt(0, 0, 0) }, "an empty wordmark must not divide by zero")
}
