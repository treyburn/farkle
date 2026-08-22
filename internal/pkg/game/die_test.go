package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
