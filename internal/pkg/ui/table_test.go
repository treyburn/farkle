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

// TestNameSortFoldsApostrophes pins the fold the name column orders by: the
// data uses U+2019 where a reader expects an ASCII apostrophe, and the two
// must not sort as different letters.
func TestNameSortFoldsApostrophes(t *testing.T) {
	assert.Equal(t, foldName("Tengri's die"), foldName("Tengri’s die"))
	assert.Equal(t, foldName("tengri's die"), foldName("Tengri’s die"),
		"and the fold is case-insensitive")
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
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6", "Joker"}, titles)

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

func TestFaceWeightSumsRepeatedFaces(t *testing.T) {
	dice, err := game.Parse([]byte(`[
  {"Id": "00000000-0000-4000-8000-00000000000e", "SideWeights": "3 4 1 1 1 1", "SideValues": "0 0 2 3 4 5", "DisplayName": "Double one die"}
]`))
	require.NoError(t, err)
	require.Len(t, dice, 1)

	// Two sides show a 1, so the face is worth both of their weights.
	assert.Equal(t, 7, FaceWeight(dice[0], game.One))
	assert.Equal(t, 0, FaceWeight(dice[0], game.Two))
	assert.Equal(t, 11, dice[0].TotalWeight())
}

func TestWeightColumnsShowRawWeights(t *testing.T) {
	dice := parseFixture(t)
	cols := WeightColumns()
	require.Len(t, cols, len(game.Pips)+3)
	assert.Equal(t, "Total", cols[len(cols)-1].Title)

	// The fixture's Weighted die is "10 1 1 1 1 1", so its 1 face is worth 10
	// of a total of 15 - reported as the integers, not as a percentage.
	var weighted game.Die
	for _, d := range dice {
		if d.Name == "Weighted die" {
			weighted = d
		}
	}
	require.NotZero(t, weighted.TotalWeight())
	assert.Equal(t, "10", cols[1].Value(weighted))
	assert.Equal(t, "15", cols[len(cols)-1].Value(weighted))
}

func TestSortByWeightMatchesSortByProb(t *testing.T) {
	// Within one die a weight and its probability differ only by the die's
	// total, but across dice with different totals they can disagree. Sorting
	// by weight must follow the weights.
	byWeight := parseFixture(t)
	SortBy(byWeight, WeightColumn(game.One), true)
	assert.Equal(t, "Weighted die", byWeight[0].Name, "weight 10 beats weight 6")

	byProb := parseFixture(t)
	SortBy(byProb, FaceColumn(game.One, game.Literal), true)
	assert.Equal(t, "Weighted die", byProb[0].Name)

	// Total orders by the denominator, which the probability columns hide.
	SortBy(byWeight, TotalColumn(), false)
	assert.Equal(t, []string{
		"Ordinary die", "Tengri’s die", "Weighted die", "Favourable die",
	}, names(byWeight))
}

func TestDefaultColumnsDropJokerWhenEffective(t *testing.T) {
	titles := func(cols []Column) []string {
		out := make([]string, len(cols))
		for i, c := range cols {
			out[i] = c.Title
		}
		return out
	}

	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6", "Joker"},
		titles(DefaultColumns(game.Literal)))
	assert.Equal(t, []string{"Name", "1", "2", "3", "4", "5", "6"},
		titles(DefaultColumns(game.Effective)),
		"Effective counts the joker into every pip, so its own column would double-count it")

	// The joker's mass really is in the pip columns it was dropped in favour of.
	tengri := parseFixture(t)[2]
	require.NotZero(t, tengri.Prob(game.Joker))
	assert.InDelta(t, tengri.Prob(game.One)+tengri.Prob(game.Joker),
		tengri.ProbBy(game.One, game.Effective), tolerance)
}

// TestSumProbCountsAJokerOnce is the arithmetic multi-select turns on. Under
// Effective a joker stands in for every pip picked, but one roll is one face:
// adding ProbEffective across the picked faces would count that joker mass
// once per face and report odds no die actually has.
func TestSumProbCountsAJokerOnce(t *testing.T) {
	tengri := parseFixture(t)[2]
	require.NotZero(t, tengri.Prob(game.Joker))
	oneFive := []game.Face{game.One, game.Five}

	assert.InDelta(t, tengri.Prob(game.One)+tengri.Prob(game.Five),
		SumProb(tengri, oneFive, game.Literal), tolerance)

	assert.InDelta(t, tengri.Prob(game.One)+tengri.Prob(game.Five)+tengri.Prob(game.Joker),
		SumProb(tengri, oneFive, game.Effective), tolerance)
	assert.Less(t, SumProb(tengri, oneFive, game.Effective),
		tengri.ProbBy(game.One, game.Effective)+tengri.ProbBy(game.Five, game.Effective),
		"summing the effective columns would double the joker")

	// Picking the joker itself alongside a pip does not count it twice either.
	assert.InDelta(t, SumProb(tengri, oneFive, game.Effective),
		SumProb(tengri, []game.Face{game.One, game.Five, game.Joker}, game.Effective), tolerance)

	// Nothing picked totals nothing, rather than everything.
	assert.Zero(t, SumProb(tengri, nil, game.Effective))
}

func TestSumColumnRanksByTheWholeSelection(t *testing.T) {
	dice := parseFixture(t)
	col := SumColumn([]game.Face{game.One, game.Five}, game.Literal)
	assert.Equal(t, "Total", col.Title)
	assert.Zero(t, col.Face, "the total is not a face, so it is never itself pickable")

	SortBy(dice, col, true)
	// Weighted die is "10 1 1 1 1 1": eleven of fifteen on the 1 and 5 faces,
	// ahead of Favourable die's twelve of eighteen.
	assert.Equal(t, "Weighted die", dice[0].Name)
	assert.Equal(t, "73.3%", col.Value(dice[0]))
	assert.Equal(t, "66.7%", col.Value(dice[1]))

	// A selection of one ranks exactly as that face's own column does.
	single := parseFixture(t)
	SortBy(single, SumColumn([]game.Face{game.One}, game.Literal), true)
	face := parseFixture(t)
	SortBy(face, FaceColumn(game.One, game.Literal), true)
	assert.Equal(t, names(face), names(single))
}

func TestMultiColumnsKeepTheTotalOnScreen(t *testing.T) {
	cols := MultiColumns([]game.Face{game.One}, game.Literal)
	assert.Equal(t, []string{"Name", "Total", "1", "2", "3", "4", "5", "6", "Joker"},
		titles(cols), "the total sits second, where a narrow terminal cannot cut it off")

	// The face columns still carry the faces the picking cursor moves over.
	assert.Equal(t, game.One, cols[2].Face)
	assert.Zero(t, cols[0].Face)
	assert.Zero(t, cols[1].Face)

	assert.Equal(t, []string{"Name", "Total", "1", "2", "3", "4", "5", "6"},
		titles(MultiColumns([]game.Face{game.One}, game.Effective)))
}

func TestEveryColumnHasATint(t *testing.T) {
	// A Column with no tint falls back to plain page colors, which would look
	// like a bug rather than a decision. Every constructor must set one.
	dice := parseFixture(t)
	sets := map[string][]Column{
		"literal":   DefaultColumns(game.Literal),
		"effective": DefaultColumns(game.Effective),
		"weights":   WeightColumns(),
		"multi":     MultiColumns([]game.Face{game.One, game.Five}, game.Literal),
		"empty":     MultiColumns(nil, game.Effective),
	}
	for name, cols := range sets {
		for _, c := range cols {
			assert.NotNil(t, c.tint, "%s column %q has no tint", name, c.Title)
			assert.NotPanics(t, func() { c.style(dice[0], true) })
		}
	}
}

// TestAnUntintedColumnStillDraws covers the fallback that keeps the assertion
// above from being the only thing standing between a hand-built Column and a
// nil call. Column is exported with exported fields, so a caller outside the
// constructors here can hand style a zero value.
func TestAnUntintedColumnStillDraws(t *testing.T) {
	d := parseFixture(t)[0]
	var bare Column
	require.Nil(t, bare.tint)

	// It still has to pick up the row background, or a selected row would break
	// into bands wherever an untinted column fell.
	assert.Equal(t, rowBG(true), bare.style(d, true).GetBackground())
	assert.Equal(t, rowBG(false), bare.style(d, false).GetBackground())
}
