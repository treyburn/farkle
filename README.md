# farkle
A TUI for interactive dice roll tables in KCD2's farkle minigame

## Data Processing
### Data Exploration
KCD2 stores it's data files under `...` with the extension of `.pak`.

A `.pak` file is simple a zipped dir of `.xml` and `.tbl` files. The `.tbl` files themselves are just "compiled" artifacts based on the raw `.xml` files that are optimized for reading by the game engine.

A `Table.pak` file contains everything we need. Once unzipped - we find a `...path/items/item.xml` file that contains what we are looking for - the dice weight tables.

### Data Processing
We can use [nushell](https://www.nushell.sh/) to make easy work of this and transform the full `items.xml` into an easier to parse subset of just the dice weights.

```nu
open ./data/items.xml | get content.0.content | where tag == 'Die' | get attributes | to json | save -f ./data/dice.json
```
