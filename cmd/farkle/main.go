// Command farkle browses the dice of KCD2's Farkle minigame in the terminal.
//
// It loads the embedded dice data at startup and shows one row per die with the
// chance of each face, sortable by any single column. Everything about how that
// looks lives in internal/pkg/ui.
package main

import (
	"fmt"
	"os"

	"go.treyburn.dev/farkle/internal/pkg/game"
	"go.treyburn.dev/farkle/internal/pkg/ui"
)

func main() {
	dice, err := game.Dice()
	// Parse reports bad entries while still returning the good ones, so a
	// partial load is worth showing as long as something survived.
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning:", err)
		os.Exit(1)
	}
	if len(dice) == 0 {
		fmt.Fprintln(os.Stderr, "no dice to show")
		os.Exit(1)
	}

	if err = ui.Run(dice); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
