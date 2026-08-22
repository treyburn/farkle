package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// press sends one keystroke and returns the model it produced.
func press(t *testing.T, m model, key string) (model, tea.Cmd) {
	t.Helper()
	var msg tea.KeyMsg
	switch key {
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		msg = tea.KeyMsg{Type: tea.KeySpace}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	next, cmd := m.Update(msg)
	return next.(model), cmd
}

// sized is a model over the real dice at a known terminal size.
func sized(t *testing.T, w, h int) model {
	t.Helper()
	dice, err := game.Dice()
	require.NoError(t, err)
	require.NotEmpty(t, dice)
	m := newModel(dice)
	m.width, m.height = w, h
	m = m.scroll()
	return m
}

func titles(cols []Column) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Title
	}
	return out
}

func TestViewCycleWrapsAndIsExhaustive(t *testing.T) {
	seen := map[view]bool{}
	v := views[0]
	for range views {
		assert.False(t, seen[v], "next() revisited %v before covering the rest", v)
		seen[v] = true
		v = v.next()
	}
	assert.Equal(t, views[0], v, "the cycle must close")

	// Every view has to describe itself; a new one added without a label would
	// otherwise show up on screen as the fallback text.
	labels := map[string]bool{}
	for _, v := range views {
		assert.NotEmpty(t, v.blurb(), "%v has no blurb", v)
		assert.NotEqual(t, "unknown view", v.String())
		assert.False(t, labels[v.String()], "duplicate label %q", v.String())
		labels[v.String()] = true
	}
}

func TestColumnsPerView(t *testing.T) {
	assert.Equal(t, []string{"Die", "1", "2", "3", "4", "5", "6", "Joker", "Total"},
		titles(columns(weights)))
	assert.Equal(t, []string{"Die", "1", "2", "3", "4", "5", "6", "Joker"},
		titles(columns(literalOdds)))
	assert.Equal(t, []string{"Die", "1", "2", "3", "4", "5", "6"},
		titles(columns(effectiveOdds)),
		"effective odds already count the joker into every pip")
}

func TestSortOnTogglesDirectionAndClamps(t *testing.T) {
	m := sized(t, 100, 30)

	// A fresh column sorts the way that column is usually asked about: numbers
	// biggest first, names ascending.
	require.Equal(t, 0, m.sort, "the table opens sorted by name")
	require.False(t, m.desc)
	m = m.sortOn(3)
	assert.Equal(t, 3, m.sort)
	assert.True(t, m.desc)

	// Re-selecting the live column reverses it instead of doing nothing.
	m = m.sortOn(3)
	assert.Equal(t, 3, m.sort)
	assert.False(t, m.desc)

	// Coming back to the name column picks ascending again rather than
	// inheriting the direction of the column just left.
	m = m.sortOn(0)
	assert.Equal(t, 0, m.sort)
	assert.False(t, m.desc)
	m = m.sortOn(3)
	assert.True(t, m.desc)

	// Out of range is ignored, so the arrow keys simply stop at the ends.
	m = m.sortOn(-1)
	assert.Equal(t, 3, m.sort)
	m = m.sortOn(len(m.cols))
	assert.Equal(t, 3, m.sort)
}

func TestSortActuallyOrdersTheDice(t *testing.T) {
	m := sized(t, 100, 30)
	m = m.setView(weights)
	total := len(m.cols) - 1
	require.Equal(t, "Total", m.cols[total].Title)

	m = m.sortOn(total)
	require.True(t, m.desc)
	for i := 1; i < len(m.dice); i++ {
		assert.GreaterOrEqual(t, m.dice[i-1].TotalWeight(), m.dice[i].TotalWeight())
	}

	m = m.sortOn(total) // reverse
	for i := 1; i < len(m.dice); i++ {
		assert.LessOrEqual(t, m.dice[i-1].TotalWeight(), m.dice[i].TotalWeight())
	}
}

func TestSetViewKeepsTheSortInRange(t *testing.T) {
	m := sized(t, 100, 30)
	m = m.setView(weights)
	m = m.sortOn(len(m.cols) - 1) // Total, which only this view has
	require.Equal(t, "Total", m.cols[m.sort].Title)

	// Effective odds is the narrowest table; the sort has to land on a column
	// that still exists rather than off the end.
	m = m.setView(effectiveOdds)
	assert.Less(t, m.sort, len(m.cols))
	assert.NotPanics(t, func() { _ = m.View() })
	assert.Equal(t, "6", m.cols[m.sort].Title)
}

func TestScrollFollowsTheCursor(t *testing.T) {
	m := sized(t, 100, chrome+3)
	require.Equal(t, 3, m.rows())
	last := len(m.dice) - 1

	m.cursor = last
	m = m.scroll()
	assert.Equal(t, last, m.cursor)
	assert.Equal(t, last-2, m.top, "the window slides down to hold the cursor")

	m.cursor = 0
	m = m.scroll()
	assert.Equal(t, 0, m.top)

	// The cursor cannot leave the dice in either direction.
	m.cursor = -5
	m = m.scroll()
	assert.Equal(t, 0, m.cursor)
	m.cursor = last + 99
	m = m.scroll()
	assert.Equal(t, last, m.cursor)

	// And the last page is a full one rather than mostly blank.
	assert.Equal(t, len(m.dice)-m.rows(), m.top)
}

func TestKeysDriveTheTable(t *testing.T) {
	m := sized(t, 100, 30)

	before := m.view
	m, _ = press(t, m, "tab")
	assert.Equal(t, before.next(), m.view, "tab advances the mode")

	m = m.sortOn(2)
	sortBefore, descBefore := m.sort, m.desc
	m, _ = press(t, m, " ")
	assert.Equal(t, sortBefore, m.sort)
	assert.Equal(t, !descBefore, m.desc, "space reverses without changing column")

	m, _ = press(t, m, "right")
	assert.Equal(t, sortBefore+1, m.sort)
	m, _ = press(t, m, "left")
	assert.Equal(t, sortBefore, m.sort)

	m, _ = press(t, m, "down")
	assert.Equal(t, 1, m.cursor)

	// The shortcuts that were removed must stay removed.
	for _, key := range []string{"n", "0", "1", "6"} {
		got, _ := press(t, m, key)
		assert.Equal(t, m.sort, got.sort, "key %q should do nothing", key)
	}

	_, cmd := press(t, m, "q")
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestWindowResizeIsHonoured(t *testing.T) {
	m := sized(t, 100, 30)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 64, Height: 20})
	m = next.(model)
	assert.Equal(t, 64, m.width)
	assert.Equal(t, 20, m.height)
	assert.Len(t, strings.Split(m.View(), "\n"), 20)
}

func TestTickAdvancesTheWordmark(t *testing.T) {
	m := sized(t, 100, 30)
	next, cmd := m.Update(tickMsg{})
	assert.Equal(t, m.phase+1, next.(model).phase)
	assert.NotNil(t, cmd, "the animation must re-arm itself or it stops after one frame")
}

// TestViewFillsTheScreenExactly is the layout contract: whatever the mode or
// the terminal, the view is exactly as many lines as the terminal is tall and
// every one of them is exactly as wide as it is wide. Anything else leaves the
// alt screen with stale cells on it.
func TestViewFillsTheScreenExactly(t *testing.T) {
	for _, w := range []int{200, 120, 80, 60, 40, 20} {
		for _, h := range []int{60, 30, 18, 13, 12, 8, 1} {
			for _, v := range views {
				m := sized(t, w, h)
				m = m.setView(v)
				lines := strings.Split(m.View(), "\n")
				require.Len(t, lines, h, "w=%d h=%d v=%v", w, h, v)
				for i, line := range lines {
					require.Equal(t, w, lipgloss.Width(line),
						"w=%d h=%d v=%v line %d", w, h, v, i)
				}
			}
		}
	}
}

func TestViewShowsWhatTheModeSays(t *testing.T) {
	strip := func(s string) string {
		var b strings.Builder
		for _, line := range strings.Split(s, "\n") {
			b.WriteString(lipgloss.NewStyle().Render(line))
		}
		return b.String()
	}
	for _, v := range views {
		m := sized(t, 120, 30)
		m = m.setView(v)
		out := strip(m.View())
		assert.Contains(t, out, v.String(), "the selector should name every mode")
		assert.Contains(t, out, v.blurb(), "the blurb should explain the live mode")
		assert.Contains(t, out, "(•) "+v.String(), "the live mode should be the filled one")
	}
}

func TestRowsShrinkWithTheTerminal(t *testing.T) {
	tall := sized(t, 100, 40)
	short := sized(t, 100, 20)
	assert.Greater(t, tall.rows(), short.rows())
	// Never zero, however cramped things get: something has to be on screen.
	assert.Equal(t, 1, sized(t, 100, 1).rows())
}

func TestPadFitsExactly(t *testing.T) {
	assert.Equal(t, "ab   ", pad("ab", 5, true))
	assert.Equal(t, "   ab", pad("ab", 5, false))
	assert.Equal(t, "abcde", pad("abcde", 5, true))
	// Cut rather than allowed to run on, which would shove the row's remaining
	// columns out of line.
	assert.Equal(t, "abcde", pad("abcdefgh", 5, true))
	// Counted in runes, not bytes: the die names carry typographic apostrophes.
	assert.Len(t, []rune(pad("Tengri’s", 5, true)), 5)
	assert.Len(t, []rune(pad("’", 5, false)), 5)
}

func TestFitSquaresOffALine(t *testing.T) {
	m := sized(t, 20, 30)
	assert.Equal(t, 20, lipgloss.Width(m.fit("short")))
	assert.Equal(t, 20, lipgloss.Width(m.fit(strings.Repeat("x", 99))))
	assert.Equal(t, 20, lipgloss.Width(m.fit("")))
}
