# Demo: sorting and switching views in the farkle TUI.
#
# Render from the repo root - every path below is relative to it:
#
#   vhs ./assets/demo.tape
#
# Needs vhs, plus the ttyd and ffmpeg it shells out to.
# Tape reference: https://github.com/charmbracelet/vhs
#
# Driven with the arrow pad keyset throughout, which is the one the footer
# opens on. Every key here has a WASD twin the demo does not press.

Output ./assets/demo.gif

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
Type "go build -o ./bin/farkle ./cmd/farkle && export PATH=$PWD/bin:$PATH && clear"
Enter
Sleep 8s
Show

Type "farkle"
Sleep 1s
Enter
Sleep 1s

# It opens sorted by name; \ flips the direction of whatever column the
# sort is on.
Type "\"
Sleep 1s
Type "\"
Sleep 1s

# ←/→ walk the sort along the columns - face 1, then face 2, where \
# turns it around and back again, then on to face 3.
Right
Sleep 1s
Right
Sleep 1s
Type "\"
Sleep 1s
Type "\"
Sleep 1s
Right
Sleep 1s

# Nudging the cursor down scrolls the table and the breadcrumbs at the
# bottom count off what is left above and below.
Down@150ms 10
Sleep 1s
Up@150ms 10
Sleep 1s

# delete cycles the view: face odds -> effective odds -> face weights and
# back, with insert walking the same cycle the other way. The sort holds
# its column across each switch, so the numbers change under it - and ←/→
# still move it once the new view is up.
Delete
Sleep 2s
Right
Sleep 1s

Delete
Sleep 1s
Left 2
Sleep 1s

Delete
Sleep 1s

# ? opens the guide: both keysets in full, and what the views and
# multi-select mean. esc backs out of it.
Type "?"
Sleep 3s
Down@150ms 8
Sleep 2s
Escape
Sleep 1s

Escape
Sleep 1s
