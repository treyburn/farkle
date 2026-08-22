package game

import "testing"

func TestProbEffectiveCountsJokers(t *testing.T) {
	tengri := parseFixture(t)[2]

	// The joker can stand in for a 1, so effectively this die rolls a usable 1
	// 2/7 of the time even though it never rolls a literal 1.
	if got := tengri.ProbEffective(One); !closeTo(got, 2.0/7.0) {
		t.Errorf("ProbEffective(One) = %v, want 2/7", got)
	}
	if got := tengri.ProbEffective(Five); !closeTo(got, 1.0/7.0+2.0/7.0) {
		t.Errorf("ProbEffective(Five) = %v, want 3/7", got)
	}
	// Asking about the joker itself must not double-count it.
	if got := tengri.ProbEffective(Joker); !closeTo(got, 2.0/7.0) {
		t.Errorf("ProbEffective(Joker) = %v, want 2/7", got)
	}
}
