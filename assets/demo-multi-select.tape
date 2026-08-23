# Demo: multi-select in the farkle TUI - picking several faces and sorting
# the table on their combined odds.
#
# Render from the repo root - every path below is relative to it:
#
#   vhs ./assets/demo-multi-select.tape
#
# Needs vhs, plus the ttyd and ffmpeg it shells out to.
# Tape reference: https://github.com/charmbracelet/vhs
#
# Driven with the arrow pad keyset throughout, which is the one the footer
# opens on. Every key here has a WASD twin the demo does not press.

Output ./assets/demo-multi-select.gif

Require go

Set Shell "bash"
Set FontSize 16
Set LineHeight 1.2
Set Padding 20
Set Framerate 24
Set TypingSpeed 150ms

# ~100x24 cells. Short on purpose: the dice table overruns a screen this
# size, which keeps the "N more below" scroll hint on camera.
Set Width 1000
Set Height 500

# Build off camera so the demo opens on a clean prompt.
Hide
Type "go build -o ./bin/farkle . && export PATH=$PWD/bin:$PATH && clear"
Enter
Sleep 8s
Show

Type "farkle"
Sleep 1s
Enter
Sleep 1s

# Start where the single column mode leaves off - sorted on face 1.
Right
Sleep 2s

# backspace hands the arrow keys and enter over to picking. The sort pins
# itself to the Total column, which opens at zero until a face is picked.
Backspace
Sleep 2s

# The cursor opens on face 1, so enter counts it straight in.
Enter
Sleep 2s

# ←/→ now walk the picking cursor rather than the sort: out to face 5,
# and the Total becomes the chance of rolling a 1 or a 5.
Right@200ms 4
Sleep 1s
Enter
Sleep 2s

# delete still switches the view. Under effective odds a joker stands in
# for any face, so the same two picks total higher than they did.
Delete
Sleep 3s

# Back down the row to add face 3 to the total.
Left@200ms 2
Sleep 1s
Enter
Sleep 2s

# backspace steps back out to the single column mode, sorting on whichever
# face the picking cursor was left on.
Backspace
Sleep 3s

Escape
Sleep 1s
