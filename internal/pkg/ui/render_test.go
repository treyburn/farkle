package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestPadFitsExactly(t *testing.T) {
	assert.Equal(t, "ab   ", pad("ab", 5, true))
	assert.Equal(t, "   ab", pad("ab", 5, false))
	assert.Equal(t, "abcde", pad("abcde", 5, true))
	// Cut rather than allowed to run on, which would shove the row's remaining
	// columns out of line.
	assert.Equal(t, "abcde", pad("abcdefgh", 5, true))
	// Counted in screen columns, not bytes: the die names carry typographic
	// apostrophes, and a wide glyph occupies two columns of its own.
	assert.Equal(t, 5, lipgloss.Width(pad("Tengri’s", 5, true)))
	assert.Equal(t, 5, lipgloss.Width(pad("’", 5, false)))
	// A wide glyph straddling the cut is dropped whole, so the cell would come
	// up a column short unless the trim is padded back out.
	assert.Equal(t, 1, lipgloss.Width(pad("漢字", 1, true)))
	assert.Equal(t, 3, lipgloss.Width(pad("漢字", 3, false)))
}

func TestFitSquaresOffALine(t *testing.T) {
	m := sized(t, 20, 30)
	assert.Equal(t, 20, lipgloss.Width(m.fit("short")))
	assert.Equal(t, 20, lipgloss.Width(m.fit(strings.Repeat("x", 99))))
	assert.Equal(t, 20, lipgloss.Width(m.fit("")))
}
