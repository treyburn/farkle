package ui

import "github.com/charmbracelet/lipgloss"

// The Catppuccin Mocha palette, https://catppuccin.com/palette. Only the
// shades this view actually uses are named; the rest are left out rather than
// carried around unused. The second group is the accent arc, kept in the
// palette's own order because ramp walks it end to end.
const (
	base     = lipgloss.Color("#1e1e2e")
	surface1 = lipgloss.Color("#45475a")
	surface2 = lipgloss.Color("#585b70")
	overlay1 = lipgloss.Color("#7f849c")
	subtext0 = lipgloss.Color("#a6adc8")
	text     = lipgloss.Color("#cdd6f4")

	rosewater = lipgloss.Color("#f5e0dc")
	flamingo  = lipgloss.Color("#f2cdcd")
	pink      = lipgloss.Color("#f5c2e7")
	mauve     = lipgloss.Color("#cba6f7")
	lavender  = lipgloss.Color("#b4befe")
	blue      = lipgloss.Color("#89b4fa")
	sapphire  = lipgloss.Color("#74c7ec")
	sky       = lipgloss.Color("#89dceb")
	teal      = lipgloss.Color("#94e2d5")
	green     = lipgloss.Color("#a6e3a1")
	yellow    = lipgloss.Color("#f9e2af")
	peach     = lipgloss.Color("#fab387")
)

// uniform is the probability of any one face on a fair six-sided die. It is
// the midpoint the heat colors read against, so a die's quirks stand out
// against what an ordinary die would do.
const uniform = 1.0 / 6.0

var (
	helpStyle = lipgloss.NewStyle().Foreground(overlay1).Background(base)
	// hintStyle is the scroll marker under the last row. It sits a shade above
	// the help, being something the reader has to act on rather than a
	// standing reminder of the keys.
	hintStyle = lipgloss.NewStyle().Foreground(subtext0).Background(base)
	// pageStyle paints the terminal behind the table, so the theme holds even
	// where the rows do not reach.
	pageStyle = lipgloss.NewStyle().Foreground(text).Background(base)
)

// rowBG is the background a row is drawn on. The cursor row gets a lifted
// surface rather than a reverse-video flip, which would fight the foreground
// colors the cells choose for themselves.
func rowBG(selected bool) lipgloss.Color {
	if selected {
		return surface1
	}
	return base
}

// headerStyle renders one column title. The sort column is inverted into a
// mauve chip so it is obvious at a glance which column orders the table.
func headerStyle(sorted bool) lipgloss.Style {
	s := lipgloss.NewStyle().Background(base)
	if sorted {
		return s.Foreground(base).Background(mauve).Bold(true)
	}
	return s.Foreground(subtext0)
}

// pickHeaderStyle renders a face column's title while multi-select is on.
//
// Two things have to be legible at once: which faces the total counts, and
// which column the arrow keys are on. The cursor takes a filled chip, the same
// shape the sorted column wears, in lavender rather than mauve so the two
// never read as the same thing; a picked column that is not under the cursor
// keeps that lavender as its text.
func pickHeaderStyle(picked, cursor bool) lipgloss.Style {
	s := lipgloss.NewStyle().Background(base)
	switch {
	case cursor:
		return s.Foreground(base).Background(lavender).Bold(true)
	case picked:
		return s.Foreground(lavender).Bold(true)
	}
	return s.Foreground(subtext0)
}

// nameStyle renders a die's name.
func nameStyle(selected bool) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lavender).Background(rowBG(selected))
}

// radioStyle renders one entry of the mode selector. The selected entry takes
// the same mauve the sorted column header uses, so "this is the live one" looks
// the same wherever it appears.
func radioStyle(on bool) lipgloss.Style {
	s := lipgloss.NewStyle().Background(base)
	if on {
		return s.Foreground(mauve).Bold(true)
	}
	return s.Foreground(overlay1)
}

// totalStyle renders a die's total weight. It is a denominator rather than a
// value to compare across dice, so it stays neutral where the face cells do
// not.
func totalStyle(selected bool) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(subtext0).Background(rowBG(selected))
}

// zeroFG is the color a face the die never rolls recedes to.
//
// It has to step up on a selected row: surface2 against the surface1 highlight
// is a contrast ratio of 1.37, which reads as an empty cell rather than a dim
// one. overlay1 on surface1 comes to 2.47, near enough to surface2's 2.46
// against the ordinary background that a zero recedes by the same amount
// whether or not its row is selected.
func zeroFG(selected bool) lipgloss.Color {
	if selected {
		return overlay1
	}
	return surface2
}

// probStyle renders a probability, tinted by how it compares to a fair die: a
// face the die never rolls recedes, one it favours warms up.
func probStyle(p float64, selected bool) lipgloss.Style {
	return heatStyle(p, uniform, selected)
}

// sumStyle renders a total over n picked faces. A sum is read against what a
// fair die would give those n faces between them rather than against a single
// face's share, so picking a second face does not turn the whole column warm.
func sumStyle(p float64, n int, selected bool) lipgloss.Style {
	return heatStyle(p, float64(n)*uniform, selected)
}

// heatStyle renders p against the mid it should be read as ordinary at: below
// runs cool, above runs warm, and nothing at all recedes.
func heatStyle(p, mid float64, selected bool) lipgloss.Style {
	s := lipgloss.NewStyle().Background(rowBG(selected))
	switch {
	case p == 0:
		return s.Foreground(zeroFG(selected))
	case p < mid-0.01:
		return s.Foreground(blue)
	case p > mid+0.01:
		return s.Foreground(peach)
	}
	return s.Foreground(text)
}

var (
	// bannerStyle is the rounded frame the title sits in.
	bannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface2).
			BorderBackground(base).
			Background(base).
			Padding(0, 1)
	// statusStyle is the right-hand summary of what the table is showing.
	statusStyle = lipgloss.NewStyle().Foreground(subtext0).Background(base)
	// blurbStyle is the line under the selector explaining the chosen view. It
	// sits a shade below statusStyle, being explanation rather than state.
	blurbStyle = lipgloss.NewStyle().Foreground(overlay1).Background(base)
	// fillStyle is blank space that keeps the banner's background unbroken.
	fillStyle = lipgloss.NewStyle().Background(base)
	// trimStyle carries no color of its own: it exists only to cut a string to
	// a column count, leaving whatever style it is rendered into untouched.
	trimStyle = lipgloss.NewStyle()

	// The guide's four voices: the page's own name, the section headings, the
	// things it is naming - keys and views - and the prose about them. They
	// take the colors the table already spends on the same jobs, so the page
	// reads as part of the program rather than a document inside it.
	guideTitleStyle = lipgloss.NewStyle().Foreground(mauve).Bold(true).Background(base)
	guideHeadStyle  = lipgloss.NewStyle().Foreground(lavender).Bold(true).Background(base)
	guideKeyStyle   = lipgloss.NewStyle().Foreground(text).Background(base)
	guideTextStyle  = lipgloss.NewStyle().Foreground(subtext0).Background(base)
)
