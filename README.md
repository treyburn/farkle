# farkle
[![license](https://img.shields.io/github/license/treyburn/farkle)](https://github.com/treyburn/farkle/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/treyburn/farkle/graph/badge.svg?token=VX7S9O4AF7)](https://codecov.io/gh/treyburn/farkle)
[![ci](https://github.com/treyburn/farkle/actions/workflows/ci.yml/badge.svg)](https://github.com/treyburn/farkle/actions)

A TUI for interactive dice roll tables in KCD2's farkle minigame.

The dice table data is pulled straight from the game files, rather than some out-of-date reddit posts.

You can sort dice based on face odds, effective odds including joker values, and overall raw weights. You can also select multiple faces and sort of combined probability.

## Sorting odds and switching to effective odds

![sorting the dice table and cycling views](./assets/demo.gif)

## Selecting multiple faces to sort on the combined odds

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for the tooling, build and test setup, and [data/README.md](data/README.md) for how the dice data is extracted and processed from the KCD2 game files.
