package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// tolerance is the slack allowed when comparing probabilities, which are sums
// of repeated float division.
const tolerance = 1e-9

// fixture mirrors the awkward shapes in the real data: a zero-weight side, a
// joker face, and a typographic apostrophe.
const fixture = `[
  {"Id": "00000000-0000-4000-8000-00000000000a", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Ordinary die"},
  {"Id": "00000000-0000-4000-8000-00000000000b", "SideWeights": "6 0 1 1 6 4", "SideValues": "0 1 2 3 4 5", "DisplayName": "Favourable die"},
  {"Id": "00000000-0000-4000-8000-00000000000c", "SideWeights": "2 1 1 1 1 1 ", "SideValues": "6 1 2 3 4 5 ", "DisplayName": "Tengri’s die"},
  {"Id": "00000000-0000-4000-8000-00000000000d", "SideWeights": "10 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Weighted die"}
]`

func parseFixture(t *testing.T) []game.Die {
	t.Helper()
	dice, err := game.Parse([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, dice, 4)
	return dice
}

// names is the display names of dice, in their current order.
func names(dice []game.Die) []string {
	out := make([]string, len(dice))
	for i, d := range dice {
		out[i] = d.Name
	}
	return out
}

func TestSortByFace(t *testing.T) {
	dice := parseFixture(t)

	SortBy(dice, FaceColumn(game.One, game.Literal), true)
	assert.Equal(t, []string{
		"Weighted die", "Favourable die", "Ordinary die", "Tengri’s die",
	}, names(dice), "Tengri’s die rolls no literal 1, so it sorts last")

	// Under Effective the joker counts, lifting Tengri's die off the bottom.
	SortBy(dice, FaceColumn(game.One, game.Effective), true)
	assert.NotEqual(t, "Tengri’s die", dice[len(dice)-1].Name)
}

func TestSortByAscending(t *testing.T) {
	dice := parseFixture(t)
	SortBy(dice, NameColumn(), false)
	assert.Equal(t, []string{
		"Favourable die", "Ordinary die", "Tengri’s die", "Weighted die",
	}, names(dice))
}

func TestFilterFoldsApostrophes(t *testing.T) {
	dice := parseFixture(t)

	assert.Len(t, Filter(dice, ""), len(dice), "an empty query keeps everything")
	assert.Len(t, Filter(dice, "DIE"), 4, "matching is case-insensitive")

	// The data uses U+2019 but a user types an ASCII apostrophe.
	got := Filter(dice, "tengri's")
	require.Len(t, got, 1)
	assert.Equal(t, "Tengri’s die", got[0].Name)

	assert.Empty(t, Filter(dice, "nonesuch"))
}

// TestDefaultColumnsCoverEveryFace pins the joker column in place. Without it a
// joker die's probability mass lands in no Literal column at all, and the row
// silently sums to less than 100%.
func TestDefaultColumnsCoverEveryFace(t *testing.T) {
	cols := DefaultColumns(game.Literal)

	titles := make([]string, len(cols))
	for i, c := range cols {
		titles[i] = c.Title
	}
	assert.Equal(t, []string{"Die", "1", "2", "3", "4", "5", "6", "Joker"}, titles)

	tengri := parseFixture(t)[2]
	var sum float64
	for _, f := range append(append([]game.Face{}, game.Pips[:]...), game.Joker) {
		sum += tengri.ProbBy(f, game.Literal)
	}
	assert.InDelta(t, 1, sum, tolerance, "every Literal face should be shown somewhere")
}

func TestFormatPct(t *testing.T) {
	assert.Equal(t, "0.0%", FormatPct(0))
	assert.Equal(t, "16.7%", FormatPct(1.0/6.0))
	assert.Equal(t, "100.0%", FormatPct(1))
}
