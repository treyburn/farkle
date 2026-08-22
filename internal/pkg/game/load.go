package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"uuid"

	"go.treyburn.dev/farkle/data"
)

// rawDie mirrors the shape of dice.json, where both the weights and the face
// values arrive as space-separated integer strings.
type rawDie struct {
	ID          string `json:"Id"`
	SideWeights string `json:"SideWeights"`
	SideValues  string `json:"SideValues"`
	DisplayName string `json:"DisplayName"`
}

// Parse decodes dice.json.
//
// A malformed entry is skipped rather than failing the whole load: the returned
// slice holds every die that parsed, and the error joins one entry per die that
// did not. Callers that want strictness can treat a non-nil error as fatal;
// callers that would rather work with the dice that survived can log it.
func Parse(data []byte) ([]Die, error) {
	var raws []rawDie
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("decode dice: %w", err)
	}

	dice := make([]Die, 0, len(raws))
	seen := make(map[uuid.UUID]string, len(raws))
	var errs []error
	for i, r := range raws {
		d, err := r.toDie()
		if err != nil {
			errs = append(errs, fmt.Errorf("entry %d (%q): %w", i, r.DisplayName, err))
			continue
		}
		// IDs are the durable identity of a die, so a collision means a corrupt
		// file rather than something to paper over.
		if prev, dup := seen[d.ID]; dup {
			errs = append(errs, fmt.Errorf("entry %d (%q): Id %s already used by %q", i, d.Name, d.ID, prev))
			continue
		}
		seen[d.ID] = d.Name
		dice = append(dice, d)
	}
	return dice, errors.Join(errs...)
}

// Dice parses the embedded dice.json.
//
// This is the normal way to get the game's dice. Parse stays exported for
// callers that have their own bytes, such as a newer dump extracted after a
// game patch.
func Dice() ([]Die, error) {
	dice, err := Parse(data.DiceJSON)
	if err != nil {
		return dice, fmt.Errorf("embedded dice.json: %w", err)
	}
	return dice, nil
}

func (r rawDie) toDie() (Die, error) {
	raw := strings.TrimSpace(r.ID)
	if raw == "" {
		return Die{}, errors.New("no Id")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return Die{}, fmt.Errorf("invalid Id %q: %w", raw, err)
	}

	weights, err := parseInts(r.SideWeights)
	if err != nil {
		return Die{}, fmt.Errorf("SideWeights: %w", err)
	}
	values, err := parseInts(r.SideValues)
	if err != nil {
		return Die{}, fmt.Errorf("SideValues: %w", err)
	}
	if len(weights) != len(values) {
		return Die{}, fmt.Errorf("have %d weights but %d values", len(weights), len(values))
	}
	if len(weights) == 0 {
		return Die{}, errors.New("no sides")
	}

	sides := make([]Side, len(weights))
	total := 0
	for i, w := range weights {
		if w < 0 {
			return Die{}, fmt.Errorf("side %d: negative weight %d", i, w)
		}
		f, err := parseFace(values[i])
		if err != nil {
			return Die{}, fmt.Errorf("side %d: %w", i, err)
		}
		sides[i] = Side{Face: f, Weight: w}
		total += w
	}
	if total == 0 {
		return Die{}, errors.New("every side weight is zero")
	}

	name := strings.TrimSpace(r.DisplayName)
	if name == "" {
		return Die{}, errors.New("no display name")
	}
	return newDie(id, name, sides, total), nil
}

// parseInts splits a space-separated integer string. The game data has stray
// trailing whitespace on some rows, which strings.Fields absorbs.
func parseInts(s string) ([]int, error) {
	fields := strings.Fields(s)
	out := make([]int, len(fields))
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("field %d: %q is not an integer", i, f)
		}
		out[i] = n
	}
	return out, nil
}
