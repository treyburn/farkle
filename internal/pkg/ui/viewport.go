package ui

// viewport is the terminal the table is drawn into and the window onto the
// dice within it: how big the screen is, which die the cursor is on, and which
// die is drawn first.
//
// It knows nothing about dice or columns. Everything it needs from the rest of
// the model - how many dice there are, how many lines the frame is spending -
// is passed in, which is what lets the scrolling be reasoned about on its own.
type viewport struct {
	// cursor is the highlighted row and top the first row drawn, so a list
	// taller than the terminal scrolls rather than being cut off.
	cursor int
	top    int
	width  int
	height int
}

// rows is how many dice fit on screen once frame lines have gone on everything
// that is not a die.
//
// At least one, however cramped the terminal: a view with no dice on it at all
// is worse than one that overruns, and [model.View] cuts the overrun off.
func (v viewport) rows(frame int) int { return max(v.height-frame, 1) }

// home puts the cursor back on the first die.
//
// Re-sorting and re-picking both move whatever die the cursor was pointing at,
// so the view returns to the top rather than leaving it on whichever die
// happened to land in that slot.
func (v viewport) home() viewport {
	v.cursor, v.top = 0, 0
	return v
}

// clamp holds the cursor inside n dice and slides the window to keep it on
// screen, frame being the lines the rows are sharing the terminal with.
//
// n of zero is a table with nothing in it, which leaves the cursor at zero
// rather than at minus one: there is no die to point at either way, and a
// negative cursor is a nonsense state for the rest of the model to read.
func (v viewport) clamp(n, frame int) viewport {
	v.cursor = min(max(v.cursor, 0), max(n-1, 0))
	rows := v.rows(frame)
	switch {
	case v.cursor < v.top:
		v.top = v.cursor
	case v.cursor >= v.top+rows:
		v.top = v.cursor - rows + 1
	}
	// The last page is a full one rather than mostly blank, so scrolling to
	// the bottom does not leave the table hanging in empty space.
	v.top = min(max(v.top, 0), max(n-rows, 0))
	return v
}
