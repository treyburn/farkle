# farkle
A TUI for interactive dice roll tables in KCD2's farkle minigame

## Data Processing
### Data Exploration
KCD2 stores it's data files under `...` with the extension of `.pak`.

A `.pak` file is simple a zipped dir of `.xml` and `.tbl` files. The `.tbl` files themselves are just "compiled" artifacts based on the raw `.xml` files that are optimized for reading by the game engine.

A `Table.pak` file contains everything we need. Once unzipped - we find a `...path/items/item.xml` file that contains what we are looking for - the dice weight tables.

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
That got us pretty close. However we were still missing in he in-game UI display name for items - as that wasn't encoded in the `item.xml` file.

Looking at the game data files - I found a `Localization` dir with an `English_xml.pak` file. Same as above - I extracted this dir and within it was a `text_ui_items.xml`.

Exploring the file - I found that it contained a mapping of `UIName` from the `item.xml` to the in-game display name (and possibly a quest text reference name - unclear what the second value means.)

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

let ui_map = (
    open ./data/text_ui_items.xml
    | get content
    | where {|row| $row.content.0.content.0.content | str contains -i 'die'}
    | get content
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
