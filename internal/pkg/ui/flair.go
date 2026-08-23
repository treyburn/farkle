package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

// blockFont draws letters three terminal rows tall out of half blocks, which
// is as close to a larger font as a terminal gets. Only the letters the
// wordmark needs are here.
//
// Three rows rather than two because two is six half-pixels short of what an E
// needs: with only four pixels of height there is nowhere to put the middle
// bar, and E comes out identical to C.
// would hide the letter shapes the layout depends on being able to read.
//
//nolint:goconst // These are pixels, not strings: naming the repeated runs
var blockFont = map[rune][3]string{
	'2': {
		"▄▀▀▄",
		"  ▄▀",
		"█▄▄▄",
	},

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
