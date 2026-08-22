// Package game holds the domain model for KCD2's Farkle minigame: the weighted
// face distribution of each die, and eventually the scoring rules and badges
// that turn a roll into points.
//
// # Layout
//
//   - face.go - the Face enum, including Joker, and the raw index decode.
//   - die.go  - Side and Die, with per-face probabilities precomputed at load.
//   - load.go - reading dice.json into []Die.
//
// # Data
//
// The dice come from the game's own item.xml, extracted by the nushell pipeline
// documented in the README and written to data/dice.json. Each die carries the
// game's UUID, so [Die.ID] identifies a die across a regenerated dice.json or a
// patch that reorders the source. The extraction collapses one duplicate pair
// (the quest and non-quest Lucky die), so display names happen to be unique
// today; nothing here relies on that.
//
// [Parse] reports malformed entries through a joined error while still
// returning the dice that did parse, so one bad entry does not cost the caller
// the whole set.
//
// # Jokers
//
// The game encodes a wild face as side value 6, which decodes to [Joker]. It is
// a first-class Face rather than a special case folded into the numbered pips,
// because how much a joker is worth is a scoring question, not a distribution
// one.
//
// That leaves two honest answers to "how often does this die roll a 5", and
// [ProbMode] selects between them. [Literal] counts only the face itself, so a
// joker die reports 0 for the pip its joker replaces and its faces sum to 1.
// [Effective] also counts jokers, since one may stand in for any pip - which
// means the faces sum past 1 by design.
//
// # Scope
//
// This package models the game and nothing else: what a die rolls and how
// often. It has no opinion on how any of it is presented.
package game
