package game

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Error(t, err, "Parse should report the five malformed entries")
	require.Len(t, dice, 2, "only the well-formed entries should survive")

	// Every rejection reason should be named, so a reader of the log can tell
	// which entry failed and why.
	assert.Contains(t, err.Error(), "Mismatched die")
	assert.Contains(t, err.Error(), "Weightless die")
	assert.Contains(t, err.Error(), "Out of range die")
	assert.Contains(t, err.Error(), "Unidentified die")
	assert.Contains(t, err.Error(), "not-a-uuid")

	assert.Equal(t, uuid.MustParse(idGood), dice[0].ID)
	assert.Equal(t, uuid.MustParse(idAlsoGood), dice[1].ID)
}

// TestParseNamesEveryRejectionReason covers the malformed shapes the existing
// fixture does not reach. dice.json is extracted from the game's own files, so
// a patch can reshape it in any of these ways; each has to be reported with
// enough detail to find the offending entry rather than merely skipped.
func TestParseNamesEveryRejectionReason(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry string
		want  string
	}{{
		name:  "non-integer weight",
		entry: `"SideWeights": "1 x 1", "SideValues": "0 1 2", "DisplayName": "Wordy weights"`,
		want:  `SideWeights: field 1: "x" is not an integer`,
	}, {
		name:  "non-integer value",
		entry: `"SideWeights": "1 1 1", "SideValues": "0 1 ?", "DisplayName": "Wordy values"`,
		want:  `SideValues: field 2: "?" is not an integer`,
	}, {
		name:  "no sides at all",
		entry: `"SideWeights": "", "SideValues": "", "DisplayName": "Sideless die"`,
		want:  "no sides",
	}, {
		name:  "negative weight",
		entry: `"SideWeights": "1 -2 1", "SideValues": "0 1 2", "DisplayName": "Debt die"`,
		want:  "side 1: negative weight -2",
	}, {
		name:  "no display name",
		entry: `"SideWeights": "1 1 1", "SideValues": "0 1 2", "DisplayName": "   "`,
		want:  "no display name",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			dice, err := Parse([]byte(`[{"Id": "` + idGood + `", ` + tc.entry + `}]`))
			require.Error(t, err)
			assert.Empty(t, dice, "a rejected entry must not reach the caller")
			assert.Contains(t, err.Error(), tc.want)
			assert.Contains(t, err.Error(), "entry 0", "the entry has to be locatable in the file")
		})
	}
}

func TestParseRejectsDuplicateIDs(t *testing.T) {
	data := `[
      {"Id": "` + idDup + `", "SideWeights": "1 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "First"},
      {"Id": "` + idDup + `", "SideWeights": "2 1 1 1 1 1", "SideValues": "0 1 2 3 4 5", "DisplayName": "Second"}
    ]`

	dice, err := Parse([]byte(data))
	require.Error(t, err, "Parse should reject the duplicate Id")
	// The first wins; a collision must not silently produce two dice that look
	// interchangeable to a caller.
	require.Len(t, dice, 1)
	assert.Equal(t, "First", dice[0].Name)
}

func TestParseRejectsMalformedJSON(t *testing.T) {
	dice, err := Parse([]byte(`{"Id": "` + idGood + `"}`))
	require.Error(t, err)
	assert.Nil(t, dice, "a decode failure yields no dice at all, unlike a bad entry")
}
