package game

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFaceStringNamesEveryFace pins the labels the table's column headers are
// built from: a pip reads as its own number rather than its internal index,
// which is one higher than the value dice.json encodes.
func TestFaceStringNamesEveryFace(t *testing.T) {
	assert.Equal(t, "1", One.String())
	assert.Equal(t, "6", Six.String())
	assert.Equal(t, "Joker", Joker.String())

	for i, f := range Pips {
		assert.Equal(t, strconv.Itoa(i+1), f.String(), "pip %d", i)
	}

	// Anything that is not a face says so rather than rendering as a stray
	// number that would read as a real column.
	assert.Equal(t, "Face(0)", Face(0).String())
	assert.Equal(t, "Face(99)", Face(99).String())
	assert.False(t, Face(0).Valid())
	assert.False(t, Face(99).Valid())
}

func TestParseFaces(t *testing.T) {
	dice := parseFixture(t)

	ordinary := dice[0]
	for _, f := range Pips {
		assert.InDelta(t, 1.0/6.0, ordinary.Prob(f), tolerance, "Ordinary die Prob(%v)", f)
	}
	assert.False(t, ordinary.HasJoker(), "Ordinary die should have no joker")

	// Face index 6 is the joker, and it lands on the side that would have been
	// a 1. Trailing whitespace in both fields must not add a phantom side.
	tengri := dice[2]
	require.Len(t, tengri.Sides, 6)
	assert.Equal(t, Joker, tengri.Sides[0].Face)
	assert.Equal(t, 7, tengri.TotalWeight())
	assert.Zero(t, tengri.Prob(One), "Tengri's die never rolls a literal 1")

	// A zero-weight side is legal; that face simply never comes up.
	assert.Zero(t, dice[1].Prob(Two), "Favourable die has a zero-weight 2")
}
