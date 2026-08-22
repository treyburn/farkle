package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tolerance is the slack allowed when comparing probabilities, which are sums
// of repeated float division.
const tolerance = 1e-9

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
	require.NoError(t, err)
	require.Len(t, dice, 4)
	return dice
}

// TestRealData guards against the shipped dice.json drifting out of shape. The
// extraction pipeline in the README collapses the one duplicate pair, so names
// are expected to be unique here even though the source XML has 44 entries.
func TestRealData(t *testing.T) {
	dice, err := Dice()
	require.NoError(t, err)
	require.NotEmpty(t, dice)

	names := make(map[string]bool, len(dice))
	for _, d := range dice {
		sum := d.Prob(Joker)
		for _, f := range Pips {
			sum += d.Prob(f)
		}
		assert.InDelta(t, 1, sum, tolerance, "%s: probabilities should sum to 1", d.Name)
		assert.False(t, names[d.Name], "%s: duplicate display name survived extraction", d.Name)
		names[d.Name] = true
	}
}
