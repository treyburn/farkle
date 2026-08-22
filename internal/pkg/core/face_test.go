package core

import "testing"

func TestParseFaces(t *testing.T) {
	dice := parseFixture(t)

	ordinary := dice[0]
	for _, f := range Pips {
		if !closeTo(ordinary.Prob(f), 1.0/6.0) {
			t.Errorf("Ordinary die Prob(%v) = %v, want 1/6", f, ordinary.Prob(f))
		}
	}
	if ordinary.HasJoker() {
		t.Error("Ordinary die should have no joker")
	}

	// Face index 6 is the joker, and it lands on the side that would have been
	// a 1. Trailing whitespace in both fields must not add a phantom side.
	tengri := dice[2]
	if len(tengri.Sides) != 6 {
		t.Fatalf("Tengri's die has %d sides, want 6", len(tengri.Sides))
	}
	if tengri.Sides[0].Face != Joker {
		t.Errorf("Tengri's die side 0 = %v, want Joker", tengri.Sides[0].Face)
	}
	if tengri.TotalWeight() != 7 {
		t.Errorf("Tengri's die total weight = %d, want 7", tengri.TotalWeight())
	}
	if !closeTo(tengri.Prob(One), 0) {
		t.Errorf("Tengri's die Prob(One) = %v, want 0", tengri.Prob(One))
	}

	// A zero-weight side is legal; that face simply never comes up.
	if got := dice[1].Prob(Two); !closeTo(got, 0) {
		t.Errorf("Favourable die Prob(Two) = %v, want 0", got)
	}
}
