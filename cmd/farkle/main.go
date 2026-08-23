// Command farkle browses the dice of KCD2's Farkle minigame in the terminal.
//
// It loads the embedded dice data at startup and shows one row per die with the
// chance of each face, sortable by any single column. Everything about how that
// looks lives in internal/pkg/ui.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"go.treyburn.dev/farkle/internal/pkg/game"
	"go.treyburn.dev/farkle/internal/pkg/ui"
)

// buildVersion reports the version recorded in the binary. Go stamps this from
// the version control tag at build time, so a release built from v0.1.0 reports
// exactly that; builds from an untagged or dirty tree report a pseudo-version.
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "unknown"
}

func main() {
	showVersion := flag.Bool("version", false, "print the farkle version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("farkle", buildVersion())
		os.Exit(0)
	}

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
