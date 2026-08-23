# farkle
[![license](https://img.shields.io/github/license/treyburn/farkle)](https://github.com/treyburn/farkle/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/treyburn/farkle/graph/badge.svg?token=VX7S9O4AF7)](https://codecov.io/gh/treyburn/farkle)
[![ci](https://github.com/treyburn/farkle/actions/workflows/ci.yml/badge.svg)](https://github.com/treyburn/farkle/actions)

A TUI for interactive dice roll tables in KCD2's farkle minigame.

The dice table data is pulled straight from the game files, rather than some out-of-date reddit posts.

You can sort dice based on face odds, effective odds including joker values, and overall raw weights. You can also select multiple faces and sort of combined probability.

### Sorting odds and switching to effective odds

Here you can see how to cycle through the different odds views and how to sort on different faces.

![sorting the dice table and cycling views](./assets/demo.gif)

### Selecting multiple faces to sort on the combined odds

Here you can see how to select multiple faces and sort on the combined probability that you will get 1 of the selected faces.

![Selecting multiple faces to sort on the combined odds](./assets/demo-multi-select.gif)

## Controls

Common key hints are listed at the bottom, and the `?` help shortcut can be used to see the fulls et of available keybindings.

Both the arrow pad and wasd gamer keys are supported modes with their own key bindings. Using one mode will update your key hints.

| Key (Alternative Key)                   | Action                                                  |
|-----------------------------------------|---------------------------------------------------------|
| `↑` `↓` (`w` `s`)                       | scroll the table up/down                                |
| `←` `→` (`a` `d`)                       | scroll the table left or right                          |
| `insert` `delete` (`shift + tab` `tab`) | cycle the view: face odds, effective odds, face weights |
| `backspace` (`x`)                       | switch to multi-select                                  |
| `enter` (`space`)                       | select face in multi-selection                          |
| `\` (`r`)                               | reverse the sort                                        |
| `home` `end` (`q` `e`)                  | jump to the top or bottom                               |
| `?`                                     | enter help menu                                         |
| `esc` or `ctrl+c`                       | quit                                                    |

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for the tooling, build and test setup, and [data/README.md](data/README.md) for how the dice data is extracted and processed from the KCD2 game files.
