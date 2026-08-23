package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWordmarkRowsStayAligned(t *testing.T) {
	m := wordmark("KCD2 FARKLE DICE")
	require.NotEmpty(t, m[0])
	// The gradient steps by character position, so the rows must be the same
	// length or the colors will not line up vertically.
	assert.Len(t, []rune(m[1]), len([]rune(m[0])))
	assert.Len(t, []rune(m[2]), len([]rune(m[1])))

	// Every glyph is square too, which is what keeps that true for any word.
	for r, g := range blockFont {
		assert.Len(t, []rune(g[1]), len([]rune(g[0])), "glyph %q", r)
		assert.Len(t, []rune(g[2]), len([]rune(g[1])), "glyph %q", r)
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
