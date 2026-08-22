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
	assert.ErrorContains(t, err, "Mismatched die")
	assert.ErrorContains(t, err, "Weightless die")
	assert.ErrorContains(t, err, "Out of range die")
	assert.ErrorContains(t, err, "Unidentified die")
	assert.ErrorContains(t, err, "not-a-uuid")

	assert.Equal(t, uuid.MustParse(idGood), dice[0].ID)
	assert.Equal(t, uuid.MustParse(idAlsoGood), dice[1].ID)
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
