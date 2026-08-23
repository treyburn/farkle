package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// frame is every line the view spends on something that is not a die: the
// banner and column header above the rows, and the help footer below them.
//
// The lines are rendered rather than counted, so the budget the rows are given
// and what actually reaches the screen cannot drift apart. Adding a line to the
// banner costs a die automatically, where a hand-counted total would quietly
// push the last row off the bottom instead.
func (m model) frame() (head, foot []string) {
	// The banner is several lines tall, so it is split apart before the
	// per-line background is applied.
	head = append(strings.Split(m.banner(), "\n"), "", m.header())
	// The blank line above the help is left blank here and filled in by
	// [model.View]: it doubles as the scroll hint, which cannot be rendered
	// until the frame's height is known.
	foot = []string{"", helpStyle.Render(m.help())}
	return head, foot
}

// frameHeight is how many lines [model.frame] takes up.
func (m model) frameHeight() int {
	head, foot := m.frame()
	return len(head) + len(foot)
}

func (m model) View() string {
	if m.guide {
		return m.guideView()
	}
	head, foot := m.frame()
	rows := m.port.rows(len(head) + len(foot))

	lines := make([]string, 0, m.port.height)
	for _, line := range head {
		lines = append(lines, m.fit(line))
	}
	end := min(m.port.top+rows, len(m.dice))
	for i := m.port.top; i < end; i++ {
		lines = append(lines, m.fit(m.row(m.dice[i], i == m.port.cursor)))
	}
	foot[0] = m.scrollHint(end)
	for _, line := range foot {
		lines = append(lines, m.fit(line))
	}

	// The alt screen holds whatever was last drawn on it, so the view has to
	// account for every line of the terminal. Short of the bottom - fewer dice
	// than there is room for - it pads; past it - a screen shorter than the
	// frame, where rows has already floored at one die - it cuts.
	for len(lines) < m.port.height {
		lines = append(lines, m.fit(""))
	}
	if len(lines) > m.port.height {
		lines = lines[:max(m.port.height, 1)]
	}
	return strings.Join(lines, "\n")
}

// help is the footer's key list: each key with what it does in brackets after
// it, so the keys line the row up and the words explain it.
//
// One keyset is listed at a time - the one the player last pressed a key from.
// Naming both halves of every pair would run off an 80-column terminal, and a
// footer that answered a WASD player in arrows would be telling them about a
// hand position they have already turned down.
//
// The order is the hands' rather than the actions': the keys that move
// something come first, then the ones that change what is on screen, then the
// way out. Both keysets and both modes are built from the one list, so the
// footer cannot reorder itself when a player changes hands.
//
// Reversing the sort is left off it. It is the one action the table already
// offers twice over - the sideways keys reverse a column they are on - and the
// room it was taking is better spent pointing at the guide, which has it along
// with everything else the footer has no space to say.
//
// The four movement keys share one entry for much the same reason. Which way an
// arrow goes is the one thing a player does not need telling, so the footer says
// only that they move something and leaves what they move to the table, where
// the cursor and the marked column are already showing it.
func (m model) help() string {
	mode, view := "backspace", "del"
	if m.keys == gamerKeys {
		mode, view = "x", "tab"
	}
	// The mode key is named for where it leads rather than what it leaves, so
	// it reads as the way through to the other mode either way round.
	toMode := "multi-select"
	if m.multi {
		toMode = "single column"
	}
	parts := []string{
		hint(m.keys.moveKeys(), "move"),
		hint(view, "change odds"),
		hint(mode, toMode),
	}
	if m.multi {
		parts = append(parts, hint(m.keys.pickKey(), "select"))
	}
	return strings.Join(append(parts, hint("?", "help"), hint("esc", "quit")), " · ")
}

// hint is one footer entry: the key, then what it does.
func hint(key, does string) string { return key + " (" + does + ")" }

// scrollHint says which way the table carries on past the screen, end being
// one past the last die drawn. It is empty when the whole table fits, which is
// the ordinary case on a tall terminal.
//
// It goes in the spacer line the footer already spends between the last row
// and the help rather than a line of its own. A hint that appeared only when
// the table overran would take a die off the screen to say so, which on a
// terminal one die short of fitting is enough to make the hint the reason it
// no longer fits.
func (m model) scrollHint(end int) string {
	var parts []string
	if m.port.top > 0 {
		parts = append(parts, fmt.Sprintf("↑ %d more above", m.port.top))
	}
	if below := len(m.dice) - end; below > 0 {
		parts = append(parts, fmt.Sprintf("↓ %d more below", below))
	}
	if len(parts) == 0 {
		return ""
	}
	return hintStyle.Render(strings.Join(parts, " · "))
}

// fit squares a line off to the terminal width: trimmed if it overruns, padded
// if it falls short, so the themed background runs edge to edge. A lipgloss
// Width would wrap the overrun onto a second line instead, which would push
// every row below it out of place.
func (m model) fit(line string) string {
	line = pageStyle.MaxWidth(m.port.width).Render(line)
	if gap := m.port.width - lipgloss.Width(line); gap > 0 {
		line += pageStyle.Render(strings.Repeat(" ", gap))
	}
	return line
}

// banner is the framed title bar: the wordmark in block letters, with a
// summary of what the table is showing on the line beneath.
func (m model) banner() string {
	mark := wordmark("KCD2 FARKLE DICE")
	status := statusStyle.Render(fmt.Sprintf("sorted by %s %s",
		m.cols[m.sort].Title, glyph(m.desc)))

	// lipgloss sizes a block by its content and padding and hangs the border
	// outside that, so the bordered block is two columns narrower than the
	// terminal and its content two narrower again.
	outer := max(m.port.width-2, 1)
	inner := max(outer-2, 1)

	// Trimmed to the content area for the same reason rows are: a wrapped
	// banner would be taller than the layout budgets for.
	trim := lipgloss.NewStyle().MaxWidth(inner)
	return bannerStyle.Width(outer).Render(strings.Join([]string{
		trim.Render(gradient(mark[0], m.phase)),
		trim.Render(gradient(mark[1], m.phase)),
		trim.Render(gradient(mark[2], m.phase)),
		// Two selectors, each with its own line of explanation under it: what
		// the numbers are, then how they are picked. Both blurbs stay up
		// whichever mode is live, so neither selector is ever the one the
		// reader has to guess at. The sort closes the banner, being the thing
		// the two of them add up to.
		trim.Render(m.viewRadio()),
		trim.Render(blurbStyle.Render(m.view.blurb())),
		trim.Render(m.modeRadio()),
		trim.Render(blurbStyle.Render(m.modeBlurb())),
		trim.Render(status),
	}, "\n"))
}

// modeBlurb says what the live selection mode does, the way [view.blurb] says
// what the live view's numbers mean.
//
// In multi-select it names the faces: the marked headers say which ones are
// counted, but only prose can say what their total is a chance of.
func (m model) modeBlurb() string {
	if !m.multi {
		return "One column orders the table; " + m.keys.sideKeys() + " moves the sort along it."
	}
	faces := m.picked.faces()
	if len(faces) == 0 {
		// Short enough to survive an 80-column terminal, which is where a
		// reader most needs telling why the column is all zeroes.
		return "No faces picked yet - " + m.keys.pickKey() + " picks the one under the cursor."
	}
	names := make([]string, len(faces))
	for i, f := range faces {
		names[i] = f.String()
	}
	s := "Total is the chance of rolling a " + strings.Join(names, " or a ") + "."
	return s
}

// viewRadio is the view selector: every view listed, the live one filled in.
// It shows what the view keys will do next as much as what is on screen now.
func (m model) viewRadio() string {
	entries := make([]string, len(views))
	for i, v := range views {
		entries[i] = radioStyle(v == m.view).Render(button(v == m.view) + v.String())
	}
	return radio(entries)
}

// modeRadio is the selection selector, the same shape as the view one because
// it answers the same kind of question - which of these is on - about how the
// table is picked and ordered. Backspace moves between the two.
func (m model) modeRadio() string {
	return radio([]string{
		radioStyle(!m.multi).Render(button(!m.multi) + "single column"),
		radioStyle(m.multi).Render(button(m.multi) + "multi-select"),
	})
}

// button is one radio button, filled where it is the live choice.
func button(on bool) string {
	if on {
		return "(•) "
	}
	return "( ) "
}

// radio spaces a row of buttons out, on the banner's own background so the
// gaps between them do not read as holes in it.
func radio(entries []string) string {
	return strings.Join(entries, fillStyle.Render("  "))
}

// glyph is the banner's sort-direction marker. The column header sticks to
// ASCII arrows, where a wide glyph would throw the columns out of line.
func glyph(desc bool) string {
	if desc {
		return "▼"
	}
	return "▲"
}

// header draws the column titles, marking the sort column with the direction
// it is ordered in and, while picking, the faces the total counts.
func (m model) header() string {
	cells := make([]string, len(m.cols))
	for i, col := range m.cols {
		title := col.Title
		if m.multi && m.picked.has(col.Face) {
			// A bullet as well as the color, so which faces are counted still
			// reads on a terminal that has thrown the palette away.
			title = "•" + title
		}
		if i == m.sort {
			title += arrow(m.desc)
		}
		cells[i] = m.headerStyle(i, col).Render(pad(title, col.Width, i == 0))
	}
	return strings.Join(cells, headerStyle(false).Render(" "))
}

// headerStyle is how column i's title is drawn. Only a face column can be
// picked, and in multi-select the sort is pinned to the total, so the two
// markings never land on the same column.
func (m model) headerStyle(i int, col Column) lipgloss.Style {
	if m.multi && col.Face.Valid() {
		return pickHeaderStyle(m.picked.has(col.Face), col.Face == m.focus)
	}
	return headerStyle(i == m.sort)
}

// row draws one die. The name leads in lavender and each probability is
// tinted by how far it sits from a fair die's.
func (m model) row(d game.Die, selected bool) string {
	cells := make([]string, len(m.cols))
	for i, col := range m.cols {
		cells[i] = col.style(d, selected).Render(pad(col.Value(d), col.Width, i == 0))
	}
	// The gaps carry the row background too, so a selected row reads as one
	// continuous band rather than a run of highlighted cells.
	gap := lipgloss.NewStyle().Background(rowBG(selected)).Render(" ")
	return strings.Join(cells, gap)
}

func arrow(desc bool) string {
	if desc {
		return " v"
	}
	return " ^"
}

// pad fits s to exactly width, left-aligned when left is set and right-aligned
// otherwise. Numbers line up better on their right edge; names on their left.
//
// Anything longer is cut rather than allowed to run on, which would shove every
// column after it out of line. The only thing that overflows in practice is a
// header whose sort arrow does not fit, and the sort column is already obvious
// from its highlight.
//
// Width here is columns on screen, not runes, the same measure [model.fit] uses
// on a whole line. A rune count would let a wide glyph in a die's name push its
// row out of line with every other one.
func pad(s string, width int, left bool) string {
	if lipgloss.Width(s) > width {
		s = trimStyle.MaxWidth(width).Render(s)
	}
	// Cutting can land a column short as well as exactly: a wide glyph
	// straddling the last column is dropped whole rather than halved, so the
	// trimmed string still has to be padded back out.
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	fill := strings.Repeat(" ", gap)
	if left {
		return s + fill
	}
	return fill + s
}
