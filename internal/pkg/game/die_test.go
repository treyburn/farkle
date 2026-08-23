package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbEffectiveCountsJokers(t *testing.T) {
	tengri := parseFixture(t)[2]

	// The joker can stand in for a 1, so effectively this die rolls a usable 1
	// 2/7 of the time even though it never rolls a literal 1.
	assert.InDelta(t, 2.0/7.0, tengri.ProbEffective(One), tolerance, "ProbEffective(One)")
	assert.InDelta(t, 3.0/7.0, tengri.ProbEffective(Five), tolerance, "ProbEffective(Five)")
	// Asking about the joker itself must not double-count it.
	assert.InDelta(t, 2.0/7.0, tengri.ProbEffective(Joker), tolerance, "ProbEffective(Joker)")
}

// TestProbByFollowsTheMode pins the switch the whole table view rests on: the
// same die and face have to give different answers under the two modes, or the
// UI's mode selector is reporting a distinction that is not there.
func TestProbByFollowsTheMode(t *testing.T) {
	tengri := parseFixture(t)[2]

	for _, f := range Pips {
		assert.InDelta(t, tengri.Prob(f), tengri.ProbBy(f, Literal), tolerance,
			"Literal should agree with Prob for %v", f)
		assert.InDelta(t, tengri.ProbEffective(f), tengri.ProbBy(f, Effective), tolerance,
			"Effective should agree with ProbEffective for %v", f)
	}

	// The two modes have to actually disagree somewhere, or the fixture is not
	// exercising the joker and this test proves nothing.
	assert.Greater(t, tengri.ProbBy(One, Effective), tengri.ProbBy(One, Literal)+tolerance,
		"a joker die should read higher under effective odds")

	// A die with no joker reads the same either way, which is what lets the UI
	// switch modes without qualifying every row.
	ordinary := parseFixture(t)[0]
	require.False(t, ordinary.HasJoker())
	for _, f := range Pips {
		assert.InDelta(t, ordinary.ProbBy(f, Literal), ordinary.ProbBy(f, Effective), tolerance,
			"%v should not move for a die with no joker", f)
	}
}

func TestProbOfAnInvalidFaceIsZero(t *testing.T) {
	d := parseFixture(t)[0]
	// The zero Face is not a face, and neither is anything past the joker. Both
	// have to answer rather than index off the end of the probability table.
	assert.Zero(t, d.Prob(Face(0)))
	assert.Zero(t, d.Prob(Face(99)))
	assert.Zero(t, d.ProbBy(Face(0), Literal))
	assert.Zero(t, d.ProbBy(Face(0), Effective))
}
