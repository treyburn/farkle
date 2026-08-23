package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The numbers here are the viewport's own, not the table's: it is handed a die
// count and a frame height and never sees either where they come from.
const (
	// testFrame is a stand-in for what model.frameHeight reports. Its actual
	// value is derived from the rendered banner, and nothing here depends on
	// what that comes to.
	testFrame = 14
	testDice  = 20
)

func TestRowsNeverFallsBelowOne(t *testing.T) {
	assert.Equal(t, 16, viewport{height: 30}.rows(testFrame))
	assert.Equal(t, 1, viewport{height: testFrame + 1}.rows(testFrame))

	// A terminal with no room left for dice still has to show one, since the
	// alternative is a screen of frame with nothing in it. View cuts off
	// whatever overruns.
	assert.Equal(t, 1, viewport{height: testFrame}.rows(testFrame))
	assert.Equal(t, 1, viewport{height: 1}.rows(testFrame))
	assert.Equal(t, 1, viewport{height: 0}.rows(testFrame))
}

func TestClampHoldsTheCursorInRange(t *testing.T) {
	v := viewport{
		height: 30, width: 100,

		// Past the end comes back to the last die, and before the start to the
		// first, so the arrow keys simply stop rather than running off.
		cursor: testDice + 99,
	}
	assert.Equal(t, testDice-1, v.clamp(testDice, testFrame).cursor)
	v.cursor = -5
	assert.Equal(t, 0, v.clamp(testDice, testFrame).cursor)

	// An empty table leaves the cursor at zero rather than at minus one: there
	// is no die to point at either way, and a negative cursor is a nonsense
	// state for the rest of the model to read.
	v.cursor = 3
	empty := v.clamp(0, testFrame)
	assert.Equal(t, 0, empty.cursor)
	assert.Equal(t, 0, empty.top)
}

func TestClampSlidesTheWindowToTheCursor(t *testing.T) {
	v := viewport{height: 30, width: 100}
	rows := v.rows(testFrame)

	// A cursor below the window pulls it down by just enough to hold it, so
	// the die the cursor is on is the last one drawn.
	v.cursor, v.top = testDice-1, 0
	got := v.clamp(testDice, testFrame)
	assert.Equal(t, testDice-rows, got.top)
	assert.Equal(t, got.top+rows-1, got.cursor, "the cursor lands on the last drawn row")

	// A cursor above the window pulls it up to sit on the cursor exactly.
	v.cursor, v.top = 2, 10
	assert.Equal(t, 2, v.clamp(testDice, testFrame).top)

	// A cursor already inside the window leaves it alone.
	v.cursor, v.top = 5, 4
	assert.Equal(t, 4, v.clamp(testDice, testFrame).top)
}

func TestClampFillsTheLastPage(t *testing.T) {
	v := viewport{height: 30, width: 100}
	rows := v.rows(testFrame)

	// Scrolled to the bottom, the last page is a full one rather than mostly
	// blank, so the table never hangs in empty space.
	v.cursor, v.top = testDice-1, testDice-1
	assert.Equal(t, testDice-rows, v.clamp(testDice, testFrame).top)

	// Fewer dice than fit leaves the window at the top rather than at a
	// negative offset.
	v.cursor, v.top = 0, 0
	assert.Equal(t, 0, v.clamp(3, testFrame).top)
}

func TestHomeReturnsToTheTop(t *testing.T) {
	v := viewport{cursor: 9, top: 5, width: 100, height: 30}.home()
	assert.Equal(t, 0, v.cursor)
	assert.Equal(t, 0, v.top)
	// The screen is not part of what home resets: re-sorting must not throw
	// away the size the terminal last reported.
	assert.Equal(t, 100, v.width)
	assert.Equal(t, 30, v.height)
}
