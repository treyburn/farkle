# farkle
[![pkg.go.dev](https://pkg.go.dev/badge/go.treyburn.dev/farkle.svg)](https://pkg.go.dev/go.treyburn.dev/farkle)
[![version](https://img.shields.io/github/v/release/treyburn/farkle)](https://github.com/treyburn/farkle/releases/latest)
[![license](https://img.shields.io/github/license/treyburn/farkle)](https://github.com/treyburn/farkle/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/treyburn/farkle/graph/badge.svg?token=VX7S9O4AF7)](https://codecov.io/gh/treyburn/farkle)
[![ci](https://github.com/treyburn/farkle/actions/workflows/ci.yml/badge.svg)](https://github.com/treyburn/farkle/actions)

A TUI for interactive dice roll tables in KCD2's farkle minigame.

The dice table data is pulled straight from the game files, rather than some out-of-date reddit posts.

You can sort dice based on face odds, effective odds including joker values, and overall raw weights. You can also select multiple faces and sort of combined probability.

## Installation

### Compile from source

If you have the [Go toolchain](https://go.dev/doc/install) on your system - then you can trivially compile this from source as it is a pure Go program with zero C dependencies.

```shell
go install go.treyburn.dev/farkle/cmd/farkle@latest
```

Just make sure you have your `$GOBIN` in your `$PATH`.

```shell
export PATH=$PATH:$(go env GOPATH)/bin
```

### Linux
This tool was developed on Linux and has been thoroughly tested for this OS on `amd64`. If you need another architecture - then open an issue!

You can download the binary from our release here: [official Linux binary](https://github.com/treyburn/farkle/releases/latest/download/farkle-linux-amd64)

### Mac
I don't think this game even runs on macOS? But either way - this TUI sure does!

You can download the binary from our release here: [official Mac binary](https://github.com/treyburn/farkle/releases/latest/download/farkle-darwin-arm64)

### Windows

> [!WARNING]
> This tool has not been tested on Windows.

This tool succeeds to compile for Windows, but I do not have a Windows machine to test it on. If you try this on Windows - then please let me know how it goes! I'd love to let others know if it works - and I'd love to fix it if it doesn't!

You can download the executable from our release here: [official Windows binary](https://github.com/treyburn/farkle/releases/latest/download/farkle-windows-amd64.exe)

## Demo

Below you can see some terminal recordings of the tool in action.

### Sorting odds and switching to effective odds

Here you can see how to cycle through the different odds views and how to sort on different faces.

![sorting the dice table and cycling views](./assets/demo.gif)

### Selecting multiple faces to sort on the combined odds

Here you can see how to select multiple faces and sort on the combined probability that you will get 1 of the selected faces.

![Selecting multiple faces to sort on the combined odds](./assets/demo-multi-select.gif)

## Controls

Common key hints are listed at the bottom, and the `?` help shortcut can be used to see the full set of available keybindings.

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

## Feedback

### Out of Date?
Have you noticed that this tool has gone out of date with the latest patch of KCD2? New dice? Dice rebalancing? [Open an issue!](https://github.com/treyburn/farkle/issues/new?template=data-update.yml)

### Something Broken?
Did the tool crash, or show you something that looks wrong? [Report a bug!](https://github.com/treyburn/farkle/issues/new?template=bug-report.yml)

### Want More?
Is there something you wish this tool did? [Request a feature!](https://github.com/treyburn/farkle/issues/new?template=feature-request.yml)

### Love this tool?
If you love this tool and want to give something? [Buy me a coffee!](https://github.com/sponsors/treyburn)
