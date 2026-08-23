package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The guide is the long answer: every key in both sets, and what the modes the
// keys switch between actually mean. The footer can only ever be a reminder -
// one keyset, one line, no room for prose - so everything that does not fit
// down there lives here instead, one ? away.
//
// It takes the whole screen rather than a panel over the table. A player who
// has stopped to read is not reading the dice at the same time, and a full page
// is one that does not have to be squeezed in around them.

// The key table's columns. The gamer key is the last one and runs to the end of
// the line, so it needs no width of its own.
const (
	doesWidth = 32
	keyWidth  = 17
)

// guideLines is the whole page, one string per line of screen.
//
// It is built rather than stored so that it can name the keys of whichever set
// the player last used, and so that it reads the views' own descriptions rather
// than keeping a second copy of them to fall out of date.
func (m model) guideLines() []string {
	var out []string
	add := func(style lipgloss.Style, format string, args ...any) {
		out = append(out, style.Render(fmt.Sprintf(format, args...)))
	}
	// Sections are separated by a blank line above their heading, which is why
	// the first one does not get one.
	head := func(s string) {
		if len(out) > 0 {
			add(guideTextStyle, "")
		}
		add(guideHeadStyle, "%s", s)
	}
	row := func(style lipgloss.Style, does, padKey, gamerKey string) {
		add(style, "  %s%s%s",
			pad(does, doesWidth, true), pad(padKey, keyWidth, true), gamerKey)
	}

	add(guideTitleStyle, "KCD2 Farkle dice")
	add(guideTextStyle, "Every die in the game, and what each one is likely to give you.")

	head("Keys")
	row(guideKeyStyle, "", "arrow pad", "wasd")
	for _, b := range bindings {
		row(guideTextStyle, b.does, keyName(b.pad), keyName(b.gamer))
	}
	// Neither of these belongs to a keyset, so both columns say the same thing.
	row(guideTextStyle, "this page", "?", "?")
	row(guideTextStyle, "quit", "esc", "esc")
	add(guideTextStyle, "")
	add(guideTextStyle, "  Both sets do the same things, so either hand can drive the table on its")
	add(guideTextStyle, "  own, and the footer follows whichever one you last pressed a key from.")
	add(guideTextStyle, "  The odds views are a cycle of three, so either key comes back round to")
	add(guideTextStyle, "  where it started.")

	head("The odds views")
	for _, v := range views {
		add(guideKeyStyle, "  %s", v.String())
		add(guideTextStyle, "    %s", v.blurb())
	}
	add(guideTextStyle, "")
	add(guideTextStyle, "  A joker stands in for whatever face you need, so under effective odds a")
	add(guideTextStyle, "  die's row can add up to past 100%%: the one joker is counted into every")
	add(guideTextStyle, "  face it could be spent on.")

	head("Multi-select")
	add(guideTextStyle, "  Single column orders the table by one column at a time. Multi-select")
	add(guideTextStyle, "  adds a Total instead: pick any faces you like and the total is the")
	add(guideTextStyle, "  chance of rolling any one of them, which is what the table is then")
	add(guideTextStyle, "  sorted on - so it answers \"which die gives me these\" rather than")
	add(guideTextStyle, "  \"which die gives me this one\".")
	add(guideTextStyle, "")
	add(guideTextStyle, "  Picked faces are marked with a bullet in the header. The filled one is")
	add(guideTextStyle, "  the face %s is pointing at, and %s picks it. The picking cursor never",
		m.keys.sideKeys(), m.keys.pickKey())
	add(guideTextStyle, "  leaves the faces, which is what keeps the sort on the total.")

	head("Reading the table")
	add(guideTextStyle, "  Cells are tinted against a fair die: warm where the die favours that")
	add(guideTextStyle, "  face, cool where it is short of one, dim where it never rolls it at")
	add(guideTextStyle, "  all. The sorted column wears a filled header with the direction it is")
	add(guideTextStyle, "  ordered in, and %s reverses it.", m.keys.reverseKey())

	return out
}

// guideView draws the guide over the whole screen, with its own footer pinned
// to the bottom line: a page that scrolls has to say how to leave it whichever
// part of it is showing.
func (m model) guideView() string {
	body := m.guideLines()
	rows := m.guideRows()
	top := min(max(m.guideTop, 0), max(len(body)-rows, 0))
	end := min(top+rows, len(body))

	lines := make([]string, 0, m.port.height)
	for _, line := range body[top:end] {
		lines = append(lines, m.fit(line))
	}
	for len(lines) < rows {
		lines = append(lines, m.fit(""))
	}
	lines = append(lines, m.fit(helpStyle.Render(m.guideFoot(top, end, len(body)))))

	// A terminal too short for even the footer gets whatever fits, the same way
	// the table does.
	if len(lines) > m.port.height {
		lines = lines[:max(m.port.height, 1)]
	}
	return strings.Join(lines, "\n")
}

// guideRows is how many lines of the page are on screen, the footer having the
// last one.
func (m model) guideRows() int { return max(m.port.height-1, 1) }

// guideFoot is the guide's own key list. It says what is left to read as well
// as how to get out, a page cut off by a short terminal giving no other sign
// that it carries on.
func (m model) guideFoot(top, end, n int) string {
	parts := []string{hint(m.keys.scrollKeys(), "scroll"), hint("esc", "back")}
	if top > 0 {
		parts = append(parts, fmt.Sprintf("↑ %d more above", top))
	}
	if below := n - end; below > 0 {
		parts = append(parts, fmt.Sprintf("↓ %d more below", below))
	}
	return strings.Join(parts, " · ")
}

// scrollGuide spends the movement keys on the page. Nothing else reaches the
// table while the guide is up: a player who stopped to read should not come
// back to a re-sorted table because the page was taller than their terminal.
func (m model) scrollGuide(act action) model {
	switch act {
	case scrollUp:
		m.guideTop--
	case scrollDown:
		m.guideTop++
	case jumpTop:
		m.guideTop = 0
	case jumpEnd:
		m.guideTop = len(m.guideLines())
	}
	return m.clampGuide()
}

// clampGuide holds the scroll inside the page, so the last screenful is a full
// one rather than mostly blank.
func (m model) clampGuide() model {
	m.guideTop = min(max(m.guideTop, 0), max(len(m.guideLines())-m.guideRows(), 0))
	return m
}
