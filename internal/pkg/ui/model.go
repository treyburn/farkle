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
// the ten-line banner, a blank line, the column header, a blank line and the
// help footer.
const chrome = 14

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

// mode is how a view counts a joker when it reports a probability. The
// weights view reports none, so it has no mode of its own and answers with the
// literal one, which is what multi-select falls back to when it takes the
// weights view over.
func (v view) mode() game.ProbMode {
	if v == effectiveOdds {
		return game.Effective
	}
	return game.Literal
}

// faceSlots bounds an array indexed by a [game.Face]. Index 0 is unused, the
// zero Face not being a valid one.
const faceSlots = int(game.Joker) + 1

// picks is the set of faces the multi-select total counts.
//
// It is an array rather than a map because a model is copied on every update:
// a map would be shared by every copy, so a selection made in one would show
// up in the ones it was branched from.
type picks [faceSlots]bool

func (p picks) has(f game.Face) bool { return f.Valid() && p[f] }

func (p picks) toggle(f game.Face) picks {
	if f.Valid() {
		p[f] = !p[f]
	}
	return p
}

// faces is the picked faces, in ascending order.
func (p picks) faces() []game.Face {
	out := make([]game.Face, 0, faceSlots)
	for f := game.One; f <= game.Joker; f++ {
		if p[f] {
			out = append(out, f)
		}
	}
	return out
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

	// multi is multi-select: several faces totalled into one column, which the
	// table is then hard sorted on. picked is the faces it counts and focus
	// indexes cols with the column the arrow keys are picking from.
	multi  bool
	picked picks
	focus  int

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
	return m.retable()
}

// columns is the table the model is currently showing.
func (m model) columns() []Column {
	if m.multi {
		return MultiColumns(m.picked.faces(), m.view.mode())
	}
	return columns(m.view)
}

// retable rebuilds the columns for the live view and mode, then re-sorts.
//
// Multi-select is hard sorted on the total: ranking dice by the picked faces
// is the whole of what the mode is for, and a sort free to wander off onto a
// single face would leave the total as decoration.
func (m model) retable() model {
	m.cols = m.columns()
	if m.multi {
		m.sort = sumColumn
		m.focus = m.nearestFace(m.focus)
	}
	m.sort = min(max(m.sort, 0), len(m.cols)-1)
	return m.resort()
}

// sumColumn is where MultiColumns puts the total: straight after the name.
const sumColumn = 1

// setView swaps the table over to v, holding the sort on the same column where
// that column still exists. Leaving the weights view drops the total column, so
// a sort on it falls back to the last column that survives.
func (m model) setView(v view) model {
	m.view = v
	m.sort = min(m.sort, len(columns(v))-1)
	m.cursor, m.top = 0, 0
	return m.retable()
}

// setMulti turns multi-select on or off, keeping the picked faces either way
// so that stepping out to read a single column and back does not cost the
// selection.
//
// Picking totals odds, and the weights view reports none, so switching the
// mode on from there moves to literal odds rather than totalling numbers that
// are not probabilities.
func (m model) setMulti(on bool) model {
	m.multi = on
	m.cursor, m.top = 0, 0
	if !on {
		// The total column is about to go, and dropping it shifts every face
		// one place left. Follow the face the picking cursor was on rather
		// than its index, which would now point at its neighbour.
		f := m.cols[m.focus].Face
		m.cols = m.columns()
		if i := m.faceIndex(f); i > 0 {
			m.sort, m.desc = i, true
		}
		return m.retable()
	}
	if m.view == weights {
		m.view = literalOdds
	}
	// The mode opens on an empty selection rather than guessing at one: which
	// faces a player is after is the question they came here to answer, and a
	// total seeded behind their back is a number they did not ask for.
	//
	// Best total first, for when they do pick: a player picking faces is
	// asking which die gives them the most, not the least.
	m.desc = true
	m.focus = m.firstFace()
	return m.retable()
}

// nextView is what tab moves to. Multi-select stays among the odds views,
// there being nothing in the weights view for it to total.
func (m model) nextView() view {
	v := m.view.next()
	if m.multi && v == weights {
		v = v.next()
	}
	return v
}

// faceIndex is where f sits in the live columns, or zero where none of them
// reports it - the joker under effective odds, or any face at all in a table
// the mode has just been switched out from under.
func (m model) faceIndex(f game.Face) int {
	if !f.Valid() {
		return 0
	}
	for i, col := range m.cols {
		if col.Face == f {
			return i
		}
	}
	return 0
}

// firstFace is the leftmost pickable column.
func (m model) firstFace() int { return m.nearestFace(0) }

// nearestFace is i if it is a face column, and the first face column after it
// otherwise. Only faces can be picked, so the cursor has to be kept off the
// name and total columns as the table is rebuilt around it.
func (m model) nearestFace(i int) int {
	for j := max(i, 0); j < len(m.cols); j++ {
		if m.cols[j].Face.Valid() {
			return j
		}
	}
	return len(m.cols) - 1
}

// focusOn walks the picking cursor one column in dir's direction, stopping at
// the ends. The name and total columns are stepped over rather than landed on.
func (m model) focusOn(dir int) model {
	for i := m.focus + dir; i >= 0 && i < len(m.cols); i += dir {
		if m.cols[i].Face.Valid() {
			m.focus = i
			break
		}
	}
	return m
}

// pick adds the focused face to the total, or takes it back out.
func (m model) pick() model {
	f := m.cols[m.focus].Face
	if !f.Valid() {
		return m
	}
	m.picked = m.picked.toggle(f)
	m.cursor, m.top = 0, 0
	return m.retable()
}

// reverse flips the sort without changing which column it is on.
func (m model) reverse() model {
	m.desc = !m.desc
	m.cursor = 0
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

	// Multi-select spends the arrow keys and space on the selection: with the
	// sort pinned to the total there is no column for them to move it to.
	case "left", "h":
		if m.multi {
			m = m.focusOn(-1)
		} else {
			m = m.sortOn(m.sort - 1)
		}
	case "right", "l":
		if m.multi {
			m = m.focusOn(1)
		} else {
			m = m.sortOn(m.sort + 1)
		}
	case " ", "enter":
		if m.multi {
			m = m.pick()
		} else {
			m = m.reverse()
		}
	case "r":
		m = m.reverse()

	case "tab":
		m = m.setView(m.nextView())
	case "m":
		m = m.setMulti(!m.multi)
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
		return pickHeaderStyle(m.picked.has(col.Face), i == m.focus)
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
