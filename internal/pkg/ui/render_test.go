package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestScrollHintPointsWhereTheTableCarriesOn(t *testing.T) {
	// Short enough that the dice do not all fit, so there is something to
	// scroll to in the first place.
	m := sized(t, 100, 24)
	rows := m.rows()
	require.Less(t, rows, len(m.dice))

	// At the top the only way on is down, and at the bottom only back up.
	assert.Equal(t, "↓ "+fmt.Sprint(len(m.dice)-rows)+" more below",
		stripANSI(m.scrollHint(rows)))
	m.port.cursor = len(m.dice) - 1
	m = m.scroll()
	assert.Equal(t, "↑ "+fmt.Sprint(m.port.top)+" more above",
		stripANSI(m.scrollHint(len(m.dice))))

	// Mid-table both ways are open, and the hint says so.
	m.port.top = 1
	assert.Equal(t, "↑ 1 more above · ↓ 1 more below",
		stripANSI(m.scrollHint(len(m.dice)-1)))
}

func TestScrollHintIsBlankWhenEverythingFits(t *testing.T) {
	m := sized(t, 100, 24)
	// The hint costs no line of its own, so a table that fits leaves the
	// footer's spacer as the blank line it already was.
	assert.Empty(t, m.scrollHint(len(m.dice)))
}

func TestScrollHintDoesNotCostARow(t *testing.T) {
	// The frame is the same height whether or not the table overruns:
	// otherwise a terminal one die short of fitting would lose a second die to
	// the hint that says so.
	tall := sized(t, 100, 200)
	short := sized(t, 100, 24)
	require.Empty(t, tall.scrollHint(len(tall.dice)))
	require.NotEmpty(t, short.scrollHint(short.rows()))
	assert.Equal(t, tall.frameHeight(), short.frameHeight())
}
