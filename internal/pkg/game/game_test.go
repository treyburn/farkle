package game

import (
	"math"
	"os"
	"testing"
)

// fixture covers the awkward shapes in the real data: trailing whitespace, a
// zero-weight side, a joker face, and a typographic apostrophe.
const fixture = `[
  {"Id": "00000000-0000-4000-8000-00000000000a", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Ordinary die"},
  {"Id": "00000000-0000-4000-8000-00000000000b", "SideWeights": "6 0 1 1 6 4", "SideValues": "0 1 2 3 4 5", "DisplayName": "Favourable die"},
  {"Id": "00000000-0000-4000-8000-00000000000c", "SideWeights": "2 1 1 1 1 1 ", "SideValues": "6 1 2 3 4 5 ", "DisplayName": "Tengri’s die"},
  {"Id": "00000000-0000-4000-8000-00000000000d", "SideWeights": "10 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Weighted die"}
]`

func parseFixture(t *testing.T) []Die {
	t.Helper()
	dice, err := Parse([]byte(fixture))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(dice) != 4 {
		t.Fatalf("got %d dice, want 4", len(dice))
	}
	return dice
}

func closeTo(got, want float64) bool { return math.Abs(got-want) < 1e-9 }

// TestRealData guards against the shipped dice.json drifting out of shape. The
// extraction pipeline in the README collapses the one duplicate pair, so names
// are expected to be unique here even though the source XML has 44 entries.
func TestRealData(t *testing.T) {
	const path = "../../../data/dice.json"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("no dice.json: %v", err)
	}

	dice, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if len(dice) == 0 {
		t.Fatal("no dice parsed")
	}

	names := make(map[string]bool, len(dice))
	for _, d := range dice {
		var sum float64
		for _, f := range Pips {
			sum += d.Prob(f)
		}
		sum += d.Prob(Joker)
		if !closeTo(sum, 1) {
			t.Errorf("%s: probabilities sum to %v, want 1", d.Name, sum)
		}
		if names[d.Name] {
			t.Errorf("%s: duplicate display name survived extraction", d.Name)
		}
		names[d.Name] = true
	}
}
