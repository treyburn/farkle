package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// page is the guide as one plain string, for asking what it says.
func page(t *testing.T, m model) string {
	t.Helper()
	require.True(t, m.guide, "the guide has to be open to be read")
	return strings.Join(m.guideLines(), "\n")
}

func TestTheGuideOpensAndCloses(t *testing.T) {
	m := sized(t, 100, 30)
	require.False(t, m.guide, "the table is what a player came for")

	m, _ = press(t, m, "?")
	assert.True(t, m.guide)
	m, _ = press(t, m, "?")
	assert.False(t, m.guide)

	// The unadvertised half of the pair does the same, both ways.
	m, _ = press(t, m, "/")
	assert.True(t, m.guide, "/ is ? without the shift")
	m, _ = press(t, m, "/")
	assert.False(t, m.guide)

	// Esc backs out of the page rather than out of the program, so reading the
	// guide cannot cost a player their table.
	m, _ = press(t, m, "?")
	m, cmd := press(t, m, "esc")
	assert.False(t, m.guide)
	assert.Nil(t, cmd, "esc closed the guide instead of quitting")

	// From the table it still quits, and ctrl+c quits from either.
	_, cmd = press(t, m, "esc")
	require.NotNil(t, cmd)
	m, _ = press(t, m, "?")
	_, cmd = press(t, m, "ctrl+c")
	require.NotNil(t, cmd)
}

// TestTheGuideKeysBelongToNeitherSet pins ? and esc as neutral: a player who
// opens the guide with ? must not find the footer has changed hands under them.
func TestTheGuideKeysBelongToNeitherSet(t *testing.T) {
	m := sized(t, 100, 30)
	m, _ = press(t, m, "w")
	require.Equal(t, gamerKeys, m.keys)

	m, _ = press(t, m, "?")
	assert.Equal(t, gamerKeys, m.keys, "? says nothing about which hand is on the keyboard")
	assert.Contains(t, page(t, m), "a/d", "the page names the live set")

	m, _ = press(t, m, "esc")
	assert.Equal(t, gamerKeys, m.keys)

	// Scrolling the page is a keyset key like any other, though, so it still
	// moves the footer over.
	m, _ = press(t, m, "?")
	m, _ = press(t, m, "down")
	assert.Equal(t, padKeys, m.keys)
	assert.Contains(t, page(t, m), "←/→")
}

// TestTheGuideLeavesTheTableAlone is why the page takes only the movement keys:
// a player reading it must come back to the table they left.
func TestTheGuideLeavesTheTableAlone(t *testing.T) {
	m := sized(t, 140, 30)
	m = m.sortOn(3)
	m, _ = press(t, m, "?")
	before := m

	for _, key := range []string{"left", "right", "enter", " ", "insert", "delete",
		"tab", "shift+tab", "backspace", "x", "\\", "r", "a", "d"} {
		got, cmd := press(t, before, key)
		assert.Nil(t, cmd)
		// Which hand is showing is allowed to move; nothing about the table is.
		assert.Equal(t, before.sort, got.sort, "key %q moved the sort", key)
		assert.Equal(t, before.desc, got.desc, "key %q reversed the sort", key)
		assert.Equal(t, before.view, got.view, "key %q changed the view", key)
		assert.Equal(t, before.multi, got.multi, "key %q changed the mode", key)
		assert.Equal(t, before.picked, got.picked, "key %q picked a face", key)
		assert.Equal(t, before.port, got.port, "key %q scrolled the table", key)
		assert.True(t, got.guide, "key %q closed the guide", key)
	}
}

func TestTheGuideScrolls(t *testing.T) {
	// Short enough that the page cannot possibly fit, which is the case that
	// needs the scrolling.
	m := sized(t, 100, 20)
	m, _ = press(t, m, "?")
	require.Greater(t, len(m.guideLines()), m.guideRows(), "the page must overrun the screen")
	last := len(m.guideLines()) - m.guideRows()

	m, _ = press(t, m, "down")
	assert.Equal(t, 1, m.guideTop)
	m, _ = press(t, m, "up")
	assert.Equal(t, 0, m.guideTop)

	// Up from the top stays put rather than running off the front of the page.
	m, _ = press(t, m, "up")
	assert.Equal(t, 0, m.guideTop)

	m, _ = press(t, m, "end")
	assert.Equal(t, last, m.guideTop, "the last screenful is a full one")
	m, _ = press(t, m, "down")
	assert.Equal(t, last, m.guideTop, "and the bottom holds")
	m, _ = press(t, m, "home")
	assert.Equal(t, 0, m.guideTop)

	// The footer says what is left in either direction, so a page cut off by a
	// short terminal does not look like the whole of it.
	m, _ = press(t, m, "down")
	foot := m.guideFoot(m.guideTop, m.guideTop+m.guideRows(), len(m.guideLines()))
	assert.Contains(t, foot, "1 more above")
	assert.Contains(t, foot, "more below")
	assert.Contains(t, foot, "esc (back)")

	// Reopening starts at the top: the page is short enough to read through,
	// and a reader coming back to it is starting again.
	m, _ = press(t, m, "?")
	m, _ = press(t, m, "?")
	assert.Equal(t, 0, m.guideTop)

	// A terminal that grows past the page pulls the scroll back to the top,
	// rather than leaving it stranded past the end.
	m, _ = press(t, m, "end")
	require.Positive(t, m.guideTop)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 200})
	assert.Equal(t, 0, next.(model).guideTop)
}

// TestTheGuideNamesEveryKey is what makes the page worth having: a key that
// exists and is not on it is a key nobody will find.
func TestTheGuideNamesEveryKey(t *testing.T) {
	m := sized(t, 100, 30)
	m, _ = press(t, m, "?")
	out := page(t, m)

	for _, b := range bindings {
		assert.Contains(t, out, b.does, "%v is not explained", b.act)
		assert.Contains(t, out, keyName(b.pad), "%v's arrow-pad key is not named", b.act)
		assert.Contains(t, out, keyName(b.gamer), "%v's wasd key is not named", b.act)
	}

	// Including the keys that are on no keyset, and the one the footer has no
	// room to mention at all.
	assert.Contains(t, out, "this page")
	assert.Contains(t, out, "quit")
	assert.Contains(t, out, "reverse the sort")
	assert.Contains(t, out, "arrow pad")
	assert.Contains(t, out, "wasd")

	// The columns line up, whatever the keys in them are called.
	assert.Contains(t, out, "  next odds view                  delete           tab")
	assert.Contains(t, out, "  previous odds view              insert           shift+tab")
}

// TestTheGuideExplainsTheModes covers the other half of the page: what the keys
// are switching between. The views describe themselves, so the page cannot
// carry a second copy of their descriptions that drifts from the banner's.
func TestTheGuideExplainsTheModes(t *testing.T) {
	m := sized(t, 100, 30)
	m, _ = press(t, m, "?")
	out := page(t, m)

	for _, v := range views {
		assert.Contains(t, out, v.String(), "the page should name every view")
		assert.Contains(t, out, v.blurb(), "and give %v's own description", v)
	}

	assert.Contains(t, out, "Multi-select")
	assert.Contains(t, out, "Total", "multi-select is the total, so the page has to say so")
	assert.Contains(t, out, "joker", "and what a joker does to the effective odds")
}

// TestTheGuideFillsTheScreenExactly is the layout contract the table already
// keeps: the page is exactly as many lines as the terminal is tall and every
// one of them exactly as wide, or the alt screen keeps stale cells.
func TestTheGuideFillsTheScreenExactly(t *testing.T) {
	for _, w := range []int{200, 120, 80, 60, 40, 20} {
		for _, h := range []int{60, 30, 18, 13, 12, 8, 2, 1} {
			m := sized(t, w, h)
			m, _ = press(t, m, "?")
			for _, top := range []int{0, 5, 999} {
				m.guideTop = top
				m = m.clampGuide()
				lines := strings.Split(m.View(), "\n")
				require.Len(t, lines, h, "w=%d h=%d top=%d", w, h, top)
				for i, line := range lines {
					require.Equal(t, w, lipgloss.Width(line),
						"w=%d h=%d top=%d line %d", w, h, top, i)
				}
			}
		}
	}
}
