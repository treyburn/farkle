package game

import (
	"testing"
	"uuid"
)

const (
	idGood     = "11111111-1111-4111-8111-111111111111"
	idAlsoGood = "22222222-2222-4222-8222-222222222222"
	idDup      = "33333333-3333-4333-8333-333333333333"
)

func TestParseSkipsBadEntries(t *testing.T) {
	data := `[
      {"Id": "` + idGood + `", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Good die"},
      {"Id": "44444444-4444-4444-8444-444444444444", "SideWeights": "1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Mismatched die"},
      {"Id": "55555555-5555-4555-8555-555555555555", "SideWeights": "0 0 0 0 0 0", "SideValues": "0 1 2 3 4 5", "DisplayName": "Weightless die"},
      {"Id": "66666666-6666-4666-8666-666666666666", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 9", "DisplayName": "Out of range die"},
      {"SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Unidentified die"},
      {"Id": "not-a-uuid", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Misidentified die"},
      {"Id": "` + idAlsoGood + `", "SideWeights": "2 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Another good die"}
    ]`

	dice, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("Parse should report the five malformed entries")
	}
	if len(dice) != 2 {
		t.Fatalf("got %d dice, want the 2 well-formed ones", len(dice))
	}
	if dice[0].ID != uuid.MustParse(idGood) || dice[1].ID != uuid.MustParse(idAlsoGood) {
		t.Errorf("kept IDs %s and %s, want %s and %s", dice[0].ID, dice[1].ID, idGood, idAlsoGood)
	}
}

func TestParseRejectsDuplicateIDs(t *testing.T) {
	data := `[
      {"Id": "` + idDup + `", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "First"},
      {"Id": "` + idDup + `", "SideWeights": "2 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Second"}
    ]`

	dice, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("Parse should reject the duplicate Id")
	}
	// The first wins; IDs key persisted state, so a collision must not silently
	// produce two rows that look identical to a caller.
	if len(dice) != 1 || dice[0].Name != "First" {
		t.Errorf("got %d dice (%+v), want only First", len(dice), dice)
	}
}
