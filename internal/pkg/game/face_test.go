package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
