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

// stripANSI flattens a rendered view into one plain string, so a test can ask
// what it says without minding how it is painted.
func stripANSI(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		b.WriteString(lipgloss.NewStyle().Render(line))
	}
	return b.String()
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
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6", "Joker", "Total"},
		titles(viewColumns(weights)))
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6", "Joker"},
		titles(viewColumns(literalOdds)))
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6"},
		titles(viewColumns(effectiveOdds)),
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

func TestMultiSelectTotalsThePickedFaces(t *testing.T) {
	m := sized(t, 140, 30)
	m = m.sortOn(m.faceIndex(game.Five)) // the 5 column
	require.Equal(t, game.Five, m.cols[m.sort].Face)

	m, _ = press(t, m, "m")
	require.True(t, m.multi)
	assert.Equal(t, []string{"Name", "Total", "1", "2", "3", "4", "5", "6", "Joker"},
		titles(m.cols))
	// Nothing is picked for the player: which faces they are after is the
	// question they came here to answer.
	assert.Empty(t, m.picked.faces())
	assert.Equal(t, sumColumn, m.sort, "multi-select is hard sorted on the total")
	assert.True(t, m.desc, "best total first")

	m.focus = game.Five
	m, _ = press(t, m, " ")
	assert.Equal(t, []game.Face{game.Five}, m.picked.faces())

	// Picking a second face widens the total rather than replacing it.
	m.focus = game.One
	m, _ = press(t, m, " ")
	assert.Equal(t, []game.Face{game.One, game.Five}, m.picked.faces())
	assert.Equal(t, sumColumn, m.sort, "picking must not move the sort off the total")

	faces := m.picked.faces()
	for i := 1; i < len(m.dice); i++ {
		assert.GreaterOrEqual(t,
			SumProb(m.dice[i-1], faces, game.Literal),
			SumProb(m.dice[i], faces, game.Literal),
			"the dice are ordered by the total, not by either face alone")
	}

	// Picking the same face again takes it back out.
	m, _ = press(t, m, " ")
	assert.Equal(t, []game.Face{game.Five}, m.picked.faces())
}

func TestMultiSelectSpendsTheArrowKeysOnFaces(t *testing.T) {
	m := sized(t, 140, 30)
	m, _ = press(t, m, "m")
	require.Equal(t, game.One, m.focus, "picking opens on the first face")

	sorted := m.sort
	m, _ = press(t, m, "right")
	assert.Equal(t, game.Two, m.focus)
	assert.Equal(t, sorted, m.sort, "the arrows move the picker, not the sort")

	// The total is not a face, so stepping left off the first face stops
	// rather than landing on it.
	m, _ = press(t, m, "left")
	m, _ = press(t, m, "left")
	assert.Equal(t, game.One, m.focus)

	// And the far end holds too.
	for range len(m.cols) + 2 {
		m, _ = press(t, m, "right")
	}
	assert.Equal(t, game.Joker, m.focus)

	// Space is spent on picking here, so reversing has its own key.
	desc := m.desc
	m, _ = press(t, m, "r")
	assert.Equal(t, !desc, m.desc)
	assert.Equal(t, sumColumn, m.sort)
}

func TestMultiSelectStaysAmongTheOddsViews(t *testing.T) {
	m := sized(t, 140, 30)
	m = m.setView(weights)
	m, _ = press(t, m, "m")
	assert.Equal(t, literalOdds, m.view, "the weights view has no odds to total")

	// Tab steps over the weights view rather than landing on a table with no
	// total column in it.
	for range len(views) + 1 {
		m, _ = press(t, m, "tab")
		require.NotEqual(t, weights, m.view)
		require.Equal(t, "Total", m.cols[m.sort].Title)
	}

	// Under effective odds the joker has no column of its own, and the picker
	// has to come off it rather than point past the end of the table.
	m = m.setView(literalOdds)
	m.focus = game.Joker
	m = m.setView(effectiveOdds)
	assert.Equal(t, game.Six, m.focus, "the picker falls back to the last face with a column")
	assert.Positive(t, m.faceIndex(m.focus), "and that face is one the table shows")
}

func TestLeavingMultiSelectKeepsThePickedFace(t *testing.T) {
	m := sized(t, 140, 30)
	m, _ = press(t, m, "m")
	m.focus = game.Six
	m, _ = press(t, m, " ")
	require.Equal(t, []game.Face{game.Six}, m.picked.faces())

	m, _ = press(t, m, "m")
	require.False(t, m.multi)
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6", "Joker"}, titles(m.cols))
	// Dropping the total shifts every face one column left; the sort follows
	// the face the picker was on rather than its old index.
	assert.Equal(t, game.Six, m.cols[m.sort].Face)

	// The selection survives the round trip, so stepping out to read one
	// column and back does not cost it.
	m, _ = press(t, m, "m")
	assert.Equal(t, []game.Face{game.Six}, m.picked.faces())
}

func TestMultiSelectViewNamesWhatItTotals(t *testing.T) {
	m := sized(t, 140, 30)
	m, _ = press(t, m, "m")

	// The mode opens empty, and has to say so rather than showing a column of
	// silent zeroes with no explanation for them.
	assert.Contains(t, stripANSI(m.View()), "No faces picked")

	m.picked = picks{}.toggle(game.One).toggle(game.Five)
	m = m.retable()
	out := stripANSI(m.View())
	assert.Contains(t, out, "(•) multi-select")
	assert.Contains(t, out, "( ) single column", "the selector shows the mode m goes back to")
	assert.Contains(t, out, "a 1 or a 5", "the blurb has to say what the total is a chance of")
	assert.Contains(t, out, "•1", "a picked column is marked in the header")
	assert.Contains(t, out, "space pick")

	m = m.setView(effectiveOdds)
	assert.Contains(t, stripANSI(m.View()), "joker counts once")

	off, _ := press(t, m, "m")
	out = stripANSI(off.View())
	assert.Contains(t, out, "(•) single column")
	assert.Contains(t, out, "( ) multi-select")
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
				for _, multi := range []bool{false, true} {
					m := sized(t, w, h)
					m = m.setView(v).setMulti(multi)
					lines := strings.Split(m.View(), "\n")
					require.Len(t, lines, h, "w=%d h=%d v=%v multi=%v", w, h, v, multi)
					for i, line := range lines {
						require.Equal(t, w, lipgloss.Width(line),
							"w=%d h=%d v=%v multi=%v line %d", w, h, v, multi, i)
					}
				}
			}
		}
	}
}

func TestViewShowsWhatTheModeSays(t *testing.T) {
	for _, v := range views {
		for _, multi := range []bool{false, true} {
			m := sized(t, 120, 30).setView(v).setMulti(multi)
			out := stripANSI(m.View())
			assert.Contains(t, out, m.view.String(), "the selector should name every view")
			assert.Contains(t, out, "(•) "+m.view.String(), "the live view should be filled in")

			// Both selectors keep their explanation up whatever the other one
			// is doing, so neither is ever left for the reader to guess at.
			assert.Contains(t, out, m.view.blurb(), "v=%v multi=%v", v, multi)
			assert.Contains(t, out, m.modeBlurb(), "v=%v multi=%v", v, multi)
			assert.NotEmpty(t, m.modeBlurb())

			// And the sort closes the banner rather than trailing a selector.
			assert.Contains(t, out, "sorted by "+m.cols[m.sort].Title)
		}
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
