package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
	s := lipgloss.NewStyle().Background(rowBG(selected))
	switch {
	case p == 0:
		return s.Foreground(zeroFG(selected))
	case p < uniform-0.01:
		return s.Foreground(blue)
	case p > uniform+0.01:
		return s.Foreground(peach)
	}
	return s.Foreground(text)
}

// ramp is the warm-to-cool arc of the Mocha accents, used to gradient the
// wordmark a letter at a time.
var ramp = []lipgloss.Color{
	rosewater,
	flamingo,
	pink,
	mauve,
	lavender,
	blue,
	sapphire,
	sky,
	teal,
	green,
	yellow,
	peach,
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
)

// blockFont draws letters three terminal rows tall out of half blocks, which
// is as close to a larger font as a terminal gets. Only the letters the
// wordmark needs are here.
//
// Three rows rather than two because two is six half-pixels short of what an E
// needs: with only four pixels of height there is nowhere to put the middle
// bar, and E comes out identical to C.
var blockFont = map[rune][3]string{
	'A': {
		"▄▀▀▄",
		"█▄▄█",
		"█  █",
	},

	'C': {
		"▄▀▀▀",
		"█   ",
		"▀▄▄▄",
	},

	'D': {
		"█▀▀▄",
		"█  █",
		"█▄▄▀",
	},

	'E': {
		"█▀▀▀",
		"█▀▀ ",
		"█▄▄▄",
	},

	'F': {
		"█▀▀▀",
		"█▀▀ ",
		"█   ",
	},

	'I': {
		"█",
		"█",
		"█",
	},

	'K': {
		"█ ▄▀",
		"██  ",
		"█ ▀▄",
	},

	'L': {
		"█   ",
		"█   ",
		"█▄▄▄",
	},

	'R': {
		"█▀▀▄",
		"█▄▄▀",
		"█ ▀▄",
	},

	' ': {
		"    ",
		"    ",
		"    ",
	},
}

// wordmark renders s as three rows of block letters, one space between glyphs.
// A letter blockFont has no glyph for is skipped, which keeps the rows the same
// length whatever it is handed.
func wordmark(s string) [3]string {
	var rows [3]string
	for _, r := range s {
		g, ok := blockFont[r]
		if !ok {
			continue
		}
		for i := range rows {
			if rows[i] != "" {
				rows[i] += " "
			}
			rows[i] += g[i]
		}
	}
	return rows
}

// gradient renders s with each character stepped along ramp, so the wordmark
// reads as one sweep of color rather than a block of it. Both rows of a
// wordmark are the same length, so stepping by character position lines the
// colors up vertically as well.
//
// phase slides the sweep along by that many characters, wrapping at the end of
// the string. Advancing it every frame walks the colors leftwards, and since
// ramp closes back on itself - peach runs into rosewater - the loop has no
// seam in it.
func gradient(s string, phase int) string {
	runes := []rune(s)
	var b strings.Builder
	for i := range runes {
		b.WriteString(lipgloss.NewStyle().
			Foreground(rampAt(i, len(runes), phase)).Background(base).Bold(true).
			Render(string(runes[i])))
	}
	return b.String()
}

// rampAt is the color of character i of an n-character run, phase steps along
// the sweep. Positions wrap at n, so raising phase walks the colors leftwards
// and back around.
func rampAt(i, n, phase int) lipgloss.Color {
	n = max(n, 1)
	return ramp[((i+phase)%n)*len(ramp)/n]
}
