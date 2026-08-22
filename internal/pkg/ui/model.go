package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// Run draws the dice table and blocks until the user quits it. Bubble Tea is an
// implementation detail of this package; a caller supplies dice and nothing
// else.
func Run(dice []game.Die) error {
	_, err := tea.NewProgram(newModel(dice), tea.WithAltScreen()).Run()
	return err
}

// chrome is the number of lines the view spends on anything that is not a die:
// the seven-line banner, a blank line, the column header, a blank line and the
// help footer.
const chrome = 11

// view is what the table's cells report. The three answer different questions,
// so this cycles rather than toggling: what the data stores, what a die
// literally rolls, and what a die is worth once jokers are counted.
type view int

const (
	// weights is the raw side weights, with the total they are read against.
	weights view = iota
	// literalOdds is the chance of rolling exactly that face.
	literalOdds
	// effectiveOdds also counts jokers, since one may stand in for any pip.
	// Rows in this mode sum past 100% by design.
	effectiveOdds
)

// views is every view, in the order tab cycles them.
var views = []view{weights, literalOdds, effectiveOdds}

func (v view) String() string {
	switch v {
	case weights:
		return "face weights"
	case literalOdds:
		return "face odds"
	case effectiveOdds:
		return "effective odds"
	}
	return "unknown view"
}

func (v view) next() view { return views[(int(v)+1)%len(views)] }

// blurb says in one line what the numbers on screen actually mean. The three
// views are easy to mix up - two of them are percentages that disagree - so the
// selector states its case rather than leaving it to the reader.
func (v view) blurb() string {
	switch v {
	case weights:
		return "Raw side weights, as the game data stores them."
	case literalOdds:
		return "How often the die rolls that face."
	case effectiveOdds:
		return "How often the die gives that face, with jokers standing in for any face."
	}
	return ""
}

// columns is the table for v. Which faces get a column is a question about the
// numbers rather than the layout, so the column sets answer it.
func columns(v view) []Column {
	switch v {
	case weights:
		return WeightColumns()
	case literalOdds:
		return DefaultColumns(game.Literal)
	case effectiveOdds:
		return DefaultColumns(game.Effective)
	}
	// Every view is named above; this is only here for the compiler.
	return DefaultColumns(game.Literal)
}

// model is the whole application state: the dice, the columns they are shown
// in, and which column currently orders them.
type model struct {
	dice []game.Die
	cols []Column
	view view

	// sort indexes cols; desc inverts that column's order.
	sort int
	desc bool

	// cursor is the highlighted row and top the first row drawn, so a list
	// taller than the terminal scrolls rather than being cut off.
	cursor int
	top    int
	width  int
	height int

	// phase is how far the wordmark's colors have travelled, in characters.
	phase int
}

func newModel(dice []game.Die) model {
	m := model{dice: dice, view: literalOdds, width: 80, height: 24}
	m.cols = columns(m.view)
	return m.resort()
}

// setView swaps the table over to v, holding the sort on the same column where
// that column still exists. Leaving the weights view drops the total column, so
// a sort on it falls back to the last column that survives.
func (m model) setView(v view) model {
	m.view = v
	m.cols = columns(v)
	m.sort = min(m.sort, len(m.cols)-1)
	m.cursor, m.top = 0, 0
	return m.resort()
}

// frameRate is how often the wordmark's colors advance. Slow enough that the
// redraw costs nothing worth measuring, quick enough to read as movement.
const frameRate = time.Second / 12

// tickMsg advances the wordmark one step along the ramp.
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(frameRate, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m.scroll(), nil
	case tickMsg:
		m.phase++
		return m, tick()
	case tea.KeyMsg:
		return m.key(msg)
	}
	return m, nil
}

// key handles a single keypress. Sorting always leaves the cursor on the first
// row, since the die it was pointing at has usually moved.
func (m model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		m.cursor--
	case "down", "j":
		m.cursor++
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(m.dice) - 1

	case "left", "h":
		m = m.sortOn(m.sort - 1)
	case "right", "l":
		m = m.sortOn(m.sort + 1)
	case "tab":
		m = m.setView(m.view.next())
	case " ", "enter":
		m.desc = !m.desc
		m.cursor = 0
		m = m.resort()
	}
	return m.scroll(), nil
}

// sortOn re-sorts by column i, or reverses it if it is already the sort
// column. Out-of-range indexes are ignored so the arrow keys stop at the ends.
func (m model) sortOn(i int) model {
	if i < 0 || i >= len(m.cols) {
		return m
	}
	if i == m.sort {
		m.desc = !m.desc
	} else {
		m.sort = i
		// Names read best ascending; a probability is nearly always asked for
		// as "which die rolls this most often".
		m.desc = i != 0
	}
	m.cursor = 0
	return m.resort()
}

func (m model) resort() model {
	SortBy(m.dice, m.cols[m.sort], m.desc)
	return m
}

// scroll clamps the cursor to the dice and the window to the cursor.
func (m model) scroll() model {
	m.cursor = min(max(m.cursor, 0), len(m.dice)-1)
	switch rows := m.rows(); {
	case m.cursor < m.top:
		m.top = m.cursor
	case m.cursor >= m.top+rows:
		m.top = m.cursor - rows + 1
	}
	m.top = min(max(m.top, 0), max(len(m.dice)-m.rows(), 0))
	return m
}

// rows is how many dice fit on screen.
func (m model) rows() int { return max(m.height-chrome, 1) }

func (m model) View() string {
	// The banner is several lines tall, so blocks are split apart before the
	// per-line background is applied.
	blocks := []string{m.banner(), "", m.header()}
	end := min(m.top+m.rows(), len(m.dice))
	for i := m.top; i < end; i++ {
		blocks = append(blocks, m.row(m.dice[i], i == m.cursor))
	}
	blocks = append(blocks, "", helpStyle.Render(
		"tab mode · ←/→ sort column · space reverse · ↑/↓ scroll · q quit"))

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
		// The sort sits right after the selector rather than flush against the
		// far edge, so everything the banner says stays in one column the eye
		// can run down.
		trim.Render(m.radio() + statusStyle.Render("  ·  ") + status),
		trim.Render(blurbStyle.Render(m.view.blurb())),
	}, "\n"))
}

// radio is the mode selector: every view listed, the live one filled in. It
// shows what tab will do next as much as what is on screen now.
func (m model) radio() string {
	entries := make([]string, len(views))
	for i, v := range views {
		mark := " "
		if v == m.view {
			mark = "•"
		}
		entries[i] = radioStyle(v == m.view).Render("(" + mark + ") " + v.String())
	}
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
// it is ordered in.
func (m model) header() string {
	cells := make([]string, len(m.cols))
	for i, col := range m.cols {
		title := col.Title
		if i == m.sort {
			title += arrow(m.desc)
		}
		cells[i] = headerStyle(i == m.sort).Render(pad(title, col.Width, i == 0))
	}
	return strings.Join(cells, headerStyle(false).Render(" "))
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
func pad(s string, width int, left bool) string {
	r := []rune(s)
	switch {
	case len(r) > width:
		return string(r[:width])
	case len(r) < width:
		gap := strings.Repeat(" ", width-len(r))
		if left {
			return s + gap
		}
		return gap + s
	}
	return s
}
