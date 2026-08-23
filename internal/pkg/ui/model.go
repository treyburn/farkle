package ui

import (
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"go.treyburn.dev/farkle/internal/pkg/game"
)

// Run draws the dice table and blocks until the user quits it. Bubble Tea is an
// implementation detail of this package; a caller supplies dice and nothing
// else.
func Run(dice []game.Die) error {
	_, err := tea.NewProgram(newModel(dice), tea.WithAltScreen()).Run()
	return err
}

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

// views is every view, in the order the view keys cycle them.
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

// step walks the cycle one view in dir's direction, wrapping at either end.
// Both directions are reachable because the view keys come as a pair.
func (v view) step(dir int) view {
	return views[((int(v)+dir)%len(views)+len(views))%len(views)]
}

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

// viewColumns is the table for v. Which faces get a column is a question about
// the numbers rather than the layout, so the column sets answer it.
func viewColumns(v view) []Column {
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
	// table is then hard sorted on. picked is the faces it counts and focus is
	// the face the arrow keys are picking from.
	//
	// focus is a face rather than a column index because the columns are
	// rebuilt under it on every view change and every pick, and an index would
	// have to be repaired against the new table each time. A face means the
	// same thing whatever columns are up.
	multi  bool
	picked picks
	focus  game.Face

	// port is the screen and how far the table is scrolled within it.
	port viewport

	// keys is the keyset the footer names: whichever one the player last
	// pressed a key from.
	keys keyset

	// guide is the help page, drawn over the whole screen in place of the
	// table, and guideTop is how far down it the reader has scrolled. The
	// table's own scroll is kept apart in port, so a player comes back to the
	// dice where they left them.
	guide    bool
	guideTop int

	// phase is how far the wordmark's colors have travelled, in characters.
	phase int
}

func newModel(dice []game.Die) model {
	m := model{
		dice: dice,
		view: literalOdds,
		// A size to render against until the first WindowSizeMsg arrives.
		port: viewport{width: 80, height: 24},
	}
	return m.retable()
}

// columns is the table the model is currently showing.
func (m model) columns() []Column {
	if m.multi {
		return MultiColumns(m.picked.faces(), m.view.mode())
	}
	return viewColumns(m.view)
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
		m.focus = m.nearestPick(m.focus)
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
	m.sort = min(m.sort, len(viewColumns(v))-1)
	m.port = m.port.home()
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
	m.port = m.port.home()
	if !on {
		// Sort on the face the picking cursor was left on, so stepping out of
		// the mode lands on the column the player was last reading.
		m.cols = m.columns()
		if i := m.faceIndex(m.focus); i > 0 {
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
	m.focus = game.One
	return m.retable()
}

// stepView is what the view keys move to, dir being which of the pair was
// pressed. Multi-select stays among the odds views, there being nothing in the
// weights view for it to total.
func (m model) stepView(dir int) view {
	v := m.view.step(dir)
	if m.multi && v == weights {
		v = v.step(dir)
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

// pickable is the faces the picking cursor can land on: every pip, and the
// joker where the live view still gives it a column of its own.
func (m model) pickable() []game.Face {
	pips := game.Pips[:]
	if m.view.mode() == game.Effective {
		// Effective odds fold the joker into every pip, so it has no column
		// here and nothing of its own left to pick.
		return pips
	}
	// Full slice expression: appending to a slice of the package-level Pips
	// array would otherwise be free to write into it.
	return append(pips[:len(pips):len(pips)], game.Joker)
}

// nearestPick is f where the live table can pick it, and the last pickable
// face otherwise - which is where a focused joker lands when the view switches
// to effective odds and its column goes.
func (m model) nearestPick(f game.Face) game.Face {
	p := m.pickable()
	if slices.Contains(p, f) {
		return f
	}
	return p[len(p)-1]
}

// focusOn walks the picking cursor one face in dir's direction, stopping at
// the ends.
func (m model) focusOn(dir int) model {
	p := m.pickable()
	i := slices.Index(p, m.focus) + dir
	if i >= 0 && i < len(p) {
		m.focus = p[i]
	}
	return m
}

// pick adds the focused face to the total, or takes it back out.
func (m model) pick() model {
	m.picked = m.picked.toggle(m.focus)
	m.port = m.port.home()
	return m.retable()
}

// reverse flips the sort without changing which column it is on.
func (m model) reverse() model {
	m.desc = !m.desc
	m.port.cursor = 0
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
		m.port.width, m.port.height = msg.Width, msg.Height
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
//
// Which keys reach which action is [bindings]' business; this only acts on what
// they resolve to.
func (m model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case key == "ctrl+c":
		return m, tea.Quit
	case isHelpKey(key):
		m.guide = !m.guide
		m.guideTop = 0
		return m, nil
	case key == "esc":
		// Esc backs out of the guide before it backs out of the program. A page
		// that covered the table and then closed the whole thing on the way out
		// would cost a player their sort for having read it.
		if m.guide {
			m.guide = false
			return m, nil
		}
		return m, tea.Quit
	}

	act, set, ok := lookup(key)
	if !ok {
		return m, nil
	}
	// The footer follows the hand: a player who reaches for one keyset is shown
	// that set from then on, rather than a list of keys they are not using.
	m.keys = set
	if m.guide {
		return m.scrollGuide(act), nil
	}

	switch act {
	case scrollUp:
		m.port.cursor--
	case scrollDown:
		m.port.cursor++
	case jumpTop:
		m.port.cursor = 0
	case jumpEnd:
		m.port.cursor = len(m.dice) - 1

	// Multi-select spends the sideways keys on the selection: with the sort
	// pinned to the total there is no column for them to move it to.
	case stepLeft:
		if m.multi {
			m = m.focusOn(-1)
		} else {
			m = m.sortOn(m.sort - 1)
		}
	case stepRight:
		if m.multi {
			m = m.focusOn(1)
		} else {
			m = m.sortOn(m.sort + 1)
		}

	// Picking is the one action with nothing to do outside multi-select; the
	// key is idle there rather than borrowed for something else, so it means
	// the same thing wherever a player presses it.
	case pickFace:
		if m.multi {
			m = m.pick()
		}
	case flipSort:
		m = m.reverse()

	case viewFwd:
		m = m.setView(m.stepView(1))
	case viewBack:
		m = m.setView(m.stepView(-1))
	case flipMode:
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
	m.port.cursor = 0
	return m.resort()
}

func (m model) resort() model {
	SortBy(m.dice, m.cols[m.sort], m.desc)
	return m
}

// scroll clamps the cursor to the dice and the window to the cursor - and the
// guide's own scroll to the guide, since a terminal that has just been resized
// can leave either of them pointing off the end.
func (m model) scroll() model {
	m.port = m.port.clamp(len(m.dice), m.frameHeight())
	return m.clampGuide()
}

// rows is how many dice fit on screen inside the live frame.
func (m model) rows() int { return m.port.rows(m.frameHeight()) }
