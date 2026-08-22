package core

import (
	"fmt"
	"strconv"
)

// Face is one face of a die.
//
// dice.json encodes faces as zero-based indices: 0..5 are the pips 1..6, and 6
// marks a joker, a wild face that may stand in for any pip.
type Face uint8

const (
	One Face = iota + 1
	Two
	Three
	Four
	Five
	Six
	Joker
)

// faceSlots bounds arrays indexed by Face. Index 0 is unused.
const faceSlots = int(Joker) + 1

// Pips are the six numbered faces, in display order.
var Pips = [...]Face{One, Two, Three, Four, Five, Six}

func (f Face) String() string {
	switch {
	case f >= One && f <= Six:
		return strconv.Itoa(int(f))
	case f == Joker:
		return "Joker"
	}
	return fmt.Sprintf("Face(%d)", uint8(f))
}

// Valid reports whether f is a face this package knows about.
func (f Face) Valid() bool { return f >= One && f <= Joker }

// parseFace converts the raw dice.json face encoding into a Face.
func parseFace(raw int) (Face, error) {
	if raw < 0 || raw > 6 {
		return 0, fmt.Errorf("face index %d out of range 0..6", raw)
	}
	return Face(raw + 1), nil
}
