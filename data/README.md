# Data

`dice.json` holds every die in the game with its face weights, and is embedded into the binary by [data.go](data.go). It is generated from the game's own data files - the `.xml` files in this directory - by the scripts described below.

## Regenerating dice.json

Both scripts need [nushell](https://www.nushell.sh/); `sync.nu` also needs `unzip` on your PATH. There are `just` recipes for both, and they can be run directly too.

### From a game install

[sync.nu](sync.nu) pulls the source xml out of a KCD2 install, drops it in this directory, and then runs the extraction:

```shell
just data-sync ~/.steam/steam/steamapps/common/KingdomComeDeliverance2

# or directly
nu ./data/sync.nu ~/.steam/steam/steamapps/common/KingdomComeDeliverance2
```

It accepts either the install root or the directory holding the `.pak`, and takes a `--locale` flag (default `English`) to pick which localization pak the display names come from.

### From the xml already here

[extract.nu](process.nu) does the extraction alone - it reads `item.xml`, `item__dlc.xml` and `text_ui_items.xml` from this directory and writes `dice.json`:

```shell
just data

# or directly
nu ./data/process.nu
```

## How This Works

The notes below are the exploration that produced the pipeline in `process.nu`. They are kept because the intermediate steps are useful for poking at the game data yourself.

### Data Exploration
KCD2 stores it's data files under `{STEAM_DIR}/steamapps/common/KingdomComeDeliverance2/Data` with the extension of `.pak`.

A `.pak` file is simple a zipped dir of `.xml` and `.tbl` files. The `.tbl` files themselves are just "compiled" artifacts based on the raw `.xml` files that are optimized for reading by the game engine.

A `Tables.pak` file contains everything we need. Once unzipped - we find a `./Libs/Tables/item/item.xml` file that contains what we are looking for: the dice weight values.

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

That last pipeline is what `process.nu` runs.
