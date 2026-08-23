# farkle
[![license](https://img.shields.io/github/license/treyburn/farkle)](https://github.com/treyburn/farkle/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/treyburn/farkle/graph/badge.svg?token=VX7S9O4AF7)](https://codecov.io/gh/treyburn/farkle)
[![ci](https://github.com/treyburn/farkle/actions/workflows/ci.yml/badge.svg)](https://github.com/treyburn/farkle/actions)

A TUI for interactive dice roll tables in KCD2's farkle minigame.

## Setup

### Required Tools for Development
The following tools are required for developing of this repo:

- [Golang](https://go.dev/dl/)
- [Nushell](https://www.nushell.sh/book/installation.html)
- [Just](https://just.systems/man/en/installation.html)
- [Golangci-lint](https://golangci-lint.run/docs/welcome/install/local/)

It can be helpful to use [mise](https://mise.jdx.dev/) for managing versioned dependencies that tend tro differ across repos like golangci-lint.

### Using Just

The following recipes are available - and you can get an up to date list by running `just` or `just help`:

```shell
> just

Available recipes:
    build         # Build for Linux (amd64)
    build-all     # Build every supported platform
    build-linux   # Build for Linux (amd64)
    build-mac     # Build for macOS (arm64)
    build-windows # Build for Windows (amd64)
    clean         # Remove the build output
    cover         # Open the coverage profile from `just test` in a browser
    fix           # Apply fixes for outdated APIs
    fmt           # Format the code
    help          # Prints all available recipes
    lint          # Lint the code, applying fixes where possible
    test          # Run the test suite with the race detector and per-package coverage
    vet           # Report suspicious constructs
```

### Updating Tool Versions
If you update the version of a tool in use (such as `golangci-lint`) then be sure to update the [mise.toml](mise.toml) as our CI will leverage that.

If you are using `mise` - then you can use the following commands:
```shell
# explicit version change
mise use golangci-lint@2.14.0

# bump everything to latest within what your pin allows
mise upgrade --bump

# check for outdated deps
mise outdated             # what's behind, and what's available 
mise latest golangci-lint # newest version that exists
```

## Data Processing
### Data Exploration
KCD2 stores it's data files under `{STEAM_DIR}/steamapps/common/KingdomComeDeliverance2/Data` with the extension of `.pak`.

A `.pak` file is simple a zipped dir of `.xml` and `.tbl` files. The `.tbl` files themselves are just "compiled" artifacts based on the raw `.xml` files that are optimized for reading by the game engine.

A `Table.pak` file contains everything we need. Once unzipped - we find a `./Libs/Tables/item/item.xml` file that contains what we are looking for: the dice weight values.

### Data Extraction
We can use [nushell](https://www.nushell.sh/) to make easy work of this and transform the full `items.xml` into an easier to parse subset of just the dice weights.

```nu
open ./data/item.xml 
| get content.0.content
| where tag == 'Die'
| get attributes
| to json
| save -f ./data/dice.json
```

### Data Processing
The above got us pretty close. However we were still missing in he in-game UI display name for items - as that wasn't encoded in the `item.xml` file.

Looking at the game data files - I found a `{STEAM_DIR}/steamapps/common/KingdomComeDeliverance2/Localization` dir with an `English_xml.pak` file. Same as above - I extracted this dir and within it was a `text_ui_items.xml`.

Exploring the file - I found that it contained an unnamed mapping of `UIName` from the `item.xml` to the in-game display name (and possibly a quest text reference name? unclear what the use is for the second value in the 3 result cell row.)

```nu
open ./data/text_ui_items.xml
| get content
| where {|row| $row.content.0.content.0.content | str contains -i 'die'} 
| get content
```

This file has a bit of an odd structure - but the first `row.0.content.0.content` maps exactly to the `UIName` value and the `row.2.content.0.content` maps to the in-game display name.

So then we can use `nu` to join these into a normalized json output:

```nu
let dice = (
    open ./data/item.xml
    | get content.0.content
    | where tag == 'Die'
    | get attributes
)

let die_names = ($dice | get UIName)

let ui_map = (
    open ./data/text_ui_items.xml
    | get content
    | get content
    | where {|row| $row.0.content.0.content in $die_names}
    | each {|row| {
        UIName: $row.0.content.0.content
        DisplayName: $row.2.content.0.content
      }}
)

$dice 
| join $ui_map UIName
| select SideWeights SideValues DisplayName
| save -f ./data/dice.json
```

Additionally, it turns out that items which originate in DLC (including dice) are stored in yet another xml file: `item__dlc.xml`. Thankfully it has the same layout as the standard `item.xml` - but we need to union those results in. Also I realized we had a single instance of a duplicated dice. `nu` to the rescue yet again.

```nu
let dice = (
    open ./data/item.xml
    | get content.0.content
    | where tag == 'Die'
    | get attributes
)

let dlc_dice = (
    open ./data/item__dlc.xml
    | get content.0.content
    | where tag == 'Die'
    | get attributes
)

# Quest item flag is useful if we need to de-dupe identical dice - like in the case with Lucky Die.
let all_dice = ($dice | append $dlc_dice | default "false" IsQuestItem)
let die_names = ($all_dice | get UIName)

let ui_map = (
    open ./data/text_ui_items.xml
    | get content
    | get content
    | where {|row| $row.0.content.0.content in $die_names}
    | each {|row| {
        UIName: $row.0.content.0.content
        DisplayName: $row.2.content.0.content
      }}
)

$all_dice
| join $ui_map UIName
# There are 2 instances of LuckyDie in the game. They are identical in states but one is a quest item and the other is not.
# Instead of having those dupliace - I choose to collapse them into a single result.
| group-by {|r| $"($r.DisplayName)|($r.Price)|($r.SideWeights | str trim)|($r.SideValues | str trim)"}
| values
| each {|g| $g | sort-by IsQuestItem | first}
| select Id SideWeights SideValues DisplayName
| save -f ./data/dice.json
```
