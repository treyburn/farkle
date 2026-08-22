package core

// Side pairs a face with the weight the game gives it. Weights are relative to
// the die's own total, not to a fixed scale, so a weight is only meaningful
// alongside Die.TotalWeight.
type Side struct {
	Face   Face
	Weight int
}

// Die is a single die from the game data with its face distribution
// precomputed. Dice are immutable once built by Parse.
type Die struct {
	// ID is the game's own UUID for the die. It survives regenerating
	// dice.json and any reordering a game patch introduces, so it is what
	// persisted state - cursor position, a user's dice inventory - should key
	// off rather than a name or a load-order index.
	ID    string
	Name  string
	Sides []Side

	total int
	probs [faceSlots]float64
}

// ProbMode selects how a joker face is counted when reporting a probability.
type ProbMode int

const (
	// Literal counts only the face itself.
	Literal ProbMode = iota
	// Effective also counts jokers, since a joker may stand in for any pip.
	// This is usually what a player means by "how often does this roll a 5".
	Effective
)

// TotalWeight is the sum of every side's weight.
func (d Die) TotalWeight() int { return d.total }

// Prob is the chance of rolling exactly f.
func (d Die) Prob(f Face) float64 {
	if !f.Valid() {
		return 0
	}
	return d.probs[f]
}

// ProbEffective is the chance of rolling something that can count as f, which
// includes jokers for any pip face.
func (d Die) ProbEffective(f Face) float64 {
	if f == Joker {
		return d.Prob(Joker)
	}
	return d.Prob(f) + d.Prob(Joker)
}

// ProbBy returns Prob or ProbEffective according to mode.
func (d Die) ProbBy(f Face, mode ProbMode) float64 {
	if mode == Effective {
		return d.ProbEffective(f)
	}
	return d.Prob(f)
}

// HasJoker reports whether any side of the die is wild.
func (d Die) HasJoker() bool { return d.probs[Joker] > 0 }

// newDie builds a Die from validated sides, precomputing the face
// probabilities so the table never recalculates during a render or sort.
func newDie(id, name string, sides []Side, total int) Die {
	d := Die{ID: id, Name: name, Sides: sides, total: total}
	for _, s := range sides {
		d.probs[s.Face] += float64(s.Weight) / float64(total)
	}
	return d
}
