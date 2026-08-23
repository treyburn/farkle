package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// keyFor is the key that runs act in set k, as the bindings table has it.
func keyFor(t *testing.T, act action, k keyset) string {
	t.Helper()
	for _, b := range bindings {
		if b.act == act {
			if k == gamerKeys {
				return b.gamer
			}
			return b.pad
		}
	}
	require.Fail(t, "unbound action", "%v is in no binding", act)
	return ""
}

// TestEveryActionHasBothKeys is the contract keys.go exists to keep: an action
// reachable from one hand position has to be reachable from the other, and no
// key may mean two things at once.
func TestEveryActionHasBothKeys(t *testing.T) {
	seen := map[string]action{}
	acts := map[action]bool{}
	for _, b := range bindings {
		assert.False(t, acts[b.act], "action %v is bound twice", b.act)
		acts[b.act] = true
		for _, key := range []string{b.pad, b.gamer} {
			if !assert.NotEmpty(t, key, "action %v is missing a key", b.act) {
				continue
			}
			if prev, dup := seen[key]; dup {
				assert.Fail(t, "duplicate binding",
					"%q runs both %v and %v", key, prev, b.act)
			}
			seen[key] = b.act
		}
	}
}

func TestLookupNamesTheActionAndTheSet(t *testing.T) {
	for _, b := range bindings {
		act, set, ok := lookup(b.pad)
		if assert.True(t, ok, "%q should be bound", b.pad) {
			assert.Equal(t, b.act, act, "key %q", b.pad)
			assert.Equal(t, padKeys, set, "key %q is on the arrow pad", b.pad)
		}
		act, set, ok = lookup(b.gamer)
		if assert.True(t, ok, "%q should be bound", b.gamer) {
			assert.Equal(t, b.act, act, "key %q", b.gamer)
			assert.Equal(t, gamerKeys, set, "key %q is a WASD one", b.gamer)
		}
	}
}

func TestLookupRejectsWhatIsNotBound(t *testing.T) {
	// Quitting belongs to neither set, so it must not be in the table: it is
	// what a player presses once they have stopped thinking about hand
	// positions, and it must not move the footer.
	for _, key := range []string{"esc", "ctrl+c"} {
		_, _, ok := lookup(key)
		assert.False(t, ok, "%q is handled apart from the keysets", key)
	}

	// The empty string is the one that matters here: viewBack has no gamer key,
	// and a keypress that stringifies to nothing must not be taken for it.
	for _, key := range []string{"", "z", "m", "j", "G"} {
		act, _, ok := lookup(key)
		assert.False(t, ok, "%q should be bound to nothing, got %v", key, act)
	}
}

// TestTheKeyNamesMatchTheBindings guards the two names that are written out
// rather than read off the table, so the footer cannot end up naming a key that
// does nothing.
func TestTheKeyNamesMatchTheBindings(t *testing.T) {
	assert.Equal(t, keyFor(t, pickFace, padKeys), padKeys.pickKey())
	// The space bar is the one key with no printable name of its own.
	assert.Equal(t, " ", keyFor(t, pickFace, gamerKeys))
	assert.Equal(t, "space", gamerKeys.pickKey())

	// The arrow pad names its sideways keys with the glyphs on them; the gamer
	// set has letters, which have to be the ones actually bound.
	assert.Equal(t, "←/→", padKeys.sideKeys())
	for _, act := range []action{stepLeft, stepRight} {
		assert.Contains(t, strings.Split(gamerKeys.sideKeys(), "/"),
			keyFor(t, act, gamerKeys), "sideKeys must name the bound %v key", act)
	}

	// The footer's one movement entry has to name all four of them, since it is
	// the only place they are mentioned at all.
	assert.Equal(t, "↑/↓/←/→", padKeys.moveKeys())
	moves := []action{scrollUp, scrollDown, stepLeft, stepRight}
	for _, act := range moves {
		assert.Contains(t, strings.Split(gamerKeys.moveKeys(), "/"),
			keyFor(t, act, gamerKeys), "moveKeys must name the bound %v key", act)
	}
	assert.Len(t, strings.Split(gamerKeys.moveKeys(), "/"), len(moves),
		"and name nothing else")

	// And the two sets never answer with the same name, or the footer would
	// not visibly change hands.
	assert.NotEqual(t, padKeys.moveKeys(), gamerKeys.moveKeys())
	assert.NotEqual(t, padKeys.sideKeys(), gamerKeys.sideKeys())
	assert.NotEqual(t, padKeys.pickKey(), gamerKeys.pickKey())
}

// TestTheFooterFollowsTheHand covers the hint switching: the keys named on
// screen are the ones the player is actually reaching for.
func TestTheFooterFollowsTheHand(t *testing.T) {
	m := sized(t, 140, 30)
	require.Equal(t, padKeys, m.keys, "the arrow pad is what an unread player tries first")
	assert.Contains(t, stripANSI(m.View()), "↑/↓/←/→ (move)")

	// One WASD key is enough to move the whole footer over.
	m, _ = press(t, m, "s")
	assert.Equal(t, gamerKeys, m.keys)
	out := stripANSI(m.View())
	assert.Contains(t, out, "w/a/s/d (move)")
	assert.Contains(t, out, "tab (change odds)")
	assert.Contains(t, out, "x (multi-select)")
	assert.NotContains(t, out, "↑/↓", "the set that was turned down should be off the screen")

	// The banner's prose moves with it rather than naming the other set.
	assert.Contains(t, out, "a/d moves the sort along it")

	m, _ = press(t, m, "x")
	require.True(t, m.multi)
	out = stripANSI(m.View())
	assert.Contains(t, out, "space (select)", "picking is named in the live set")
	assert.Contains(t, out, "space picks the one under the cursor")

	// And back again on the first arrow-pad key.
	m, _ = press(t, m, "up")
	assert.Equal(t, padKeys, m.keys)
	out = stripANSI(m.View())
	assert.Contains(t, out, "↑/↓/←/→ (move)")
	assert.Contains(t, out, "del (change odds)")
	assert.Contains(t, out, "enter (select)")
	assert.NotContains(t, out, "w/a/s/d")
}

// TestTheFooterReadsInOrder pins the footer whole. The order is a deliberate
// one - what moves, then what changes, then the way out - and reading it back
// as a single string is the only way a reordering shows up as a failure rather
// than as four still-passing substring checks.
func TestTheFooterReadsInOrder(t *testing.T) {
	m := sized(t, 140, 30)

	m.keys = padKeys
	assert.Equal(t,
		"↑/↓/←/→ (move) · del (change odds) · "+
			"backspace (multi-select) · ? (help) · esc (quit)",
		m.setMulti(false).help())
	assert.Equal(t,
		"↑/↓/←/→ (move) · del (change odds) · "+
			"backspace (single column) · enter (select) · ? (help) · esc (quit)",
		m.setMulti(true).help())

	m.keys = gamerKeys
	assert.Equal(t,
		"w/a/s/d (move) · tab (change odds) · "+
			"x (multi-select) · ? (help) · esc (quit)",
		m.setMulti(false).help())
	assert.Equal(t,
		"w/a/s/d (move) · tab (change odds) · "+
			"x (single column) · space (select) · ? (help) · esc (quit)",
		m.setMulti(true).help())

	// Reversing is off the footer, but it is still a key: the guide is where
	// it is spelled out now.
	for _, k := range []keyset{padKeys, gamerKeys} {
		assert.NotContains(t, model{keys: k}.help(), k.reverseKey()+" (")
	}
}

// TestQuittingLeavesTheFooterAlone is the other half of quitting being outside
// the keysets: the model is not disturbed on the way out.
func TestQuittingLeavesTheFooterAlone(t *testing.T) {
	m := sized(t, 140, 30)
	m, _ = press(t, m, "w")
	require.Equal(t, gamerKeys, m.keys)

	got, cmd := press(t, m, "esc")
	require.NotNil(t, cmd)
	assert.Equal(t, m, got, "quitting changes nothing on screen")
}
