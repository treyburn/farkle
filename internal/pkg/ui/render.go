package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

func (m model) View() string {
	// The banner is several lines tall, so blocks are split apart before the
	// per-line background is applied.
	blocks := []string{m.banner(), "", m.header()}
	end := min(m.top+m.rows(), len(m.dice))
	for i := m.top; i < end; i++ {
		blocks = append(blocks, m.row(m.dice[i], i == m.cursor))
	}
	blocks = append(blocks, "", helpStyle.Render(m.help()))

	var lines []string
	for _, block := range blocks {
		for _, line := range strings.Split(block, "\n") {
			lines = append(lines, m.fit(line))
		}
	}
	// The alt screen holds whatever was last drawn on it, so the view has to
	// account for every line of the terminal. Short of the bottom - fewer dice
	// than there is room for - it pads; past it - a screen shorter than the
	// chrome, where rows() has already floored at one die - it cuts.
	for len(lines) < m.height {
		lines = append(lines, m.fit(""))
	}
	if len(lines) > m.height {
		lines = lines[:max(m.height, 1)]
	}
	return strings.Join(lines, "\n")
}

// help is the footer's key list. It names what the keys do here rather than
// everything they might do, since the arrow keys and space change hands when
// multi-select takes over.
func (m model) help() string {
	if m.multi {
		return "m single column · tab view · ←/→ face · space pick · r reverse · ↑/↓ scroll · q quit"
	}
	return "m multi-select · tab view · ←/→ sort column · space reverse · ↑/↓ scroll · q quit"
}

// fit squares a line off to the terminal width: trimmed if it overruns, padded
// if it falls short, so the themed background runs edge to edge. A lipgloss
// Width would wrap the overrun onto a second line instead, which would push
// every row below it out of place.
func (m model) fit(line string) string {
	line = pageStyle.MaxWidth(m.width).Render(line)
	if gap := m.width - lipgloss.Width(line); gap > 0 {
		line += pageStyle.Render(strings.Repeat(" ", gap))
	}
	return line
}

// banner is the framed title bar: the wordmark in block letters, with a
// summary of what the table is showing on the line beneath.
func (m model) banner() string {
	mark := wordmark("FARKLE DICE")
	status := statusStyle.Render(fmt.Sprintf("sorted by %s %s",
		m.cols[m.sort].Title, glyph(m.desc)))

	// lipgloss sizes a block by its content and padding and hangs the border
	// outside that, so the frame is two columns narrower than the terminal.
	frame := max(m.width-2, 1)
	inner := max(frame-2, 1)

	// Trimmed to the content area for the same reason rows are: a wrapped
	// banner would be taller than the layout budgets for.
	trim := lipgloss.NewStyle().MaxWidth(inner)
	return bannerStyle.Width(frame).Render(strings.Join([]string{
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
		return "One column orders the table; ←/→ moves the sort along it."
	}
	faces := m.picked.faces()
	if len(faces) == 0 {
		// Short enough to survive an 80-column terminal, which is where a
		// reader most needs telling why the column is all zeroes.
		return "No faces picked yet - space picks the one under the cursor."
	}
	names := make([]string, len(faces))
	for i, f := range faces {
		names[i] = f.String()
	}
	s := "Total is the chance of rolling a " + strings.Join(names, " or a ") + "."
	if m.view.mode() == game.Effective {
		s += " A joker counts once, not once per face."
	}
	return s
}

// viewRadio is the view selector: every view listed, the live one filled in.
// It shows what tab will do next as much as what is on screen now.
func (m model) viewRadio() string {
	entries := make([]string, len(views))
	for i, v := range views {
		entries[i] = radioStyle(v == m.view).Render(button(v == m.view) + v.String())
	}
	return radio(entries)
}

// modeRadio is the selection selector, the same shape as the view one because
// it answers the same kind of question - which of these is on - about how the
// table is picked and ordered. m moves between the two.
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
