package game

import "testing"

func TestParseSkipsBadEntries(t *testing.T) {
	const data = `[
      {"Id": "a", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Good die"},
      {"Id": "b", "SideWeights": "1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Mismatched die"},
      {"Id": "c", "SideWeights": "0 0 0 0 0 0", "SideValues": "0 1 2 3 4 5", "DisplayName": "Weightless die"},
      {"Id": "d", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 9", "DisplayName": "Out of range die"},
      {"SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Unidentified die"},
      {"Id": "e", "SideWeights": "2 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Another good die"}
    ]`

	dice, err := Parse([]byte(data))
	if err == nil {
		t.Fatal("Parse should report the four malformed entries")
	}
	if len(dice) != 2 {
		t.Fatalf("got %d dice, want the 2 well-formed ones", len(dice))
	}
	if dice[0].ID != "a" || dice[1].ID != "e" {
		t.Errorf("kept IDs %q and %q, want a and e", dice[0].ID, dice[1].ID)
	}
}

func TestParseRejectsDuplicateIDs(t *testing.T) {
	const data = `[
      {"Id": "dup", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "First"},
      {"Id": "dup", "SideWeights": "2 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Second"}
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
