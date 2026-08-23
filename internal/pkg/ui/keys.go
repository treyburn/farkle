package ui

// This file is the whole of what the keyboard means: which keys run which
// action, and which of the two keysets each one belongs to.
//
// Every action answers to two keys, so the table can be driven one-handed from
// either end of the keyboard - the arrow pad and the keys around it, or the
// WASD cluster and its neighbors. Both are listed in the one table below
// rather than split into a switch apiece, so a key bound in one set cannot
// quietly drift away from its opposite number in the other.

// keyset is which hand position the table is being driven from.
type keyset int

const (
	// padKeys is the arrow pad and the keys within reach of it. It is the
	// default because it is the set a player who has read nothing will try.
	padKeys keyset = iota
	// gamerKeys is WASD and its neighbors.
	gamerKeys
)

// action is one thing a keypress does, named apart from the keys that run it
// because two keys run each of them.
type action int

const (
	scrollUp action = iota
	scrollDown
	jumpTop
	jumpEnd
	stepLeft
	stepRight
	pickFace
	flipSort
	viewFwd
	viewBack
	flipMode
)

// binding is one action, the key that runs it in each set, and what the guide
// calls it. The label lives here rather than with the page that prints it so
// that an action added to the table arrives with its own explanation.
type binding struct {
	act        action
	pad, gamer string
	does       string
}

// bindings is every action a key can reach, in the order the guide lists them
// and the footer summarises them: what moves around the table first, then what
// changes the table itself.
//
// Quitting and the guide are not among them: esc, ctrl+c and ? belong to
// neither set, being what a player presses when they have stopped thinking
// about which set they are in.
var bindings = []binding{
	{stepLeft, "left", "a", "move left"},
	{stepRight, "right", "d", "move right"},
	{scrollUp, "up", "w", "scroll up"},
	{scrollDown, "down", "s", "scroll down"},
	{jumpTop, "home", "q", "jump to the top"},
	{jumpEnd, "end", "e", "jump to the bottom"},
	// Delete and insert walk the views in either direction, forward on the
	// lower of the two keys so that the pair reads the way the arrow pad below
	// it does. Tab pairs with shift+tab for the same reason it always has.
	{viewFwd, "delete", "tab", "next odds view"},
	{viewBack, "insert", "shift+tab", "previous odds view"},
	{flipSort, "\\", "r", "reverse the sort"},
	{flipMode, "backspace", "x", "multi-select on and off"},
	{pickFace, "enter", " ", "pick the face under the cursor"},
}

// lookup is what key does and which set it came from, or false where the key is
// bound to nothing.
func lookup(key string) (action, keyset, bool) {
	for _, b := range bindings {
		switch {
		case key == b.pad:
			return b.act, padKeys, true
		// An action with no key in a set leaves that field empty, which must not
		// match a keypress that stringifies to nothing.
		case b.gamer != "" && key == b.gamer:
			return b.act, gamerKeys, true
		}
	}
	return 0, padKeys, false
}

// isHelpKey says whether key opens or closes the guide.
//
// Two keys, and only one of them advertised: ? is what a player will try, and /
// is the same physical key without the shift, which is a kindness to anyone who
// reaches for it in a hurry rather than something the footer needs to spend
// room on.
func isHelpKey(key string) bool { return key == "?" || key == "/" }

// keyName is how a key is written on screen. The bindings match on the strings
// Bubble Tea reports, and a few of them are not what a player would call the
// key they pressed.
func keyName(key string) string {
	switch key {
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	case " ":
		return "space"
	}
	return key
}

// sideKeys names the pair that moves along a row - the sort column in single
// column, the picking cursor in multi-select.
//
// It, [keyset.scrollKeys] and [keyset.pickKey] exist so that the footer, the
// guide and the banner's prose name the same keys as each other. The rest of
// the footer's names are written out where it is built: they appear once, and
// spelling "ins/del" out of "insert" and "delete" would cost more than it
// saves.
func (k keyset) sideKeys() string {
	if k == gamerKeys {
		return "a/d"
	}
	return "←/→"
}

// moveKeys names all four movement keys at once, which is how the footer asks
// for them: what each one does is the one thing about this program a player can
// be relied on to guess, so the footer spends a single entry on the lot rather
// than a line and a half saying which way is up.
func (k keyset) moveKeys() string {
	if k == gamerKeys {
		return "w/a/s/d"
	}
	return "↑/↓/←/→"
}

// scrollKeys names the pair that scrolls: the dice on the table, the page in
// the guide.
func (k keyset) scrollKeys() string {
	if k == gamerKeys {
		return "w/s"
	}
	return "↑/↓"
}

// reverseKey names the key that reverses the sort. The footer has no room for
// it any more, so the guide's prose is the only place it is spelled out.
func (k keyset) reverseKey() string {
	if k == gamerKeys {
		return "r"
	}
	return "\\"
}

// pickKey names the key that picks the face under the cursor.
func (k keyset) pickKey() string {
	if k == gamerKeys {
		return "space"
	}
	return "enter"
}
