#!/usr/bin/env nu

# Pull the source xml files out of a KCD2 install and rebuild dice.json.
#
# The game ships its data as .pak files, which are plain zip archives. We only
# need two of them:
#
#   {game_dir}/Data/Tables.pak        -> Libs/Tables/item/item.xml, item__dlc.xml
#   {game_dir}/Localization/*_xml.pak -> text_ui_items.xml
#
# Requires `unzip` on PATH.
def main [
    game_dir: string # KCD2 install dir, e.g. ~/.steam/steam/steamapps/common/KingdomComeDeliverance2
    --data-dir: string # Directory to extract into. Defaults to this script's directory.
    --locale: string = "English" # Localization pak to read display names from.
] {
    let dir = ($data_dir | default $env.FILE_PWD)
    let game = ($game_dir | path expand)

    if not ($game | path exists) {
        error make {msg: $"no such directory: ($game)"}
    }
    if (which unzip | is-empty) {
        error make {msg: "unzip is required but was not found on PATH"}
    }

    # Tolerate being pointed at either the install root or the dir holding the pak.
    let tables = (find_pak $game ["Data/Table*.pak" "Table*.pak"])
    let localization = (find_pak $game [$"Localization/($locale)_xml.pak" $"($locale)_xml.pak"])

    print $"extracting from ($tables)"
    ^unzip -o -j $tables "*/item.xml" "*/item__dlc.xml" -d $dir | ignore

    print $"extracting from ($localization)"
    ^unzip -o -j $localization "*text_ui_items.xml" -d $dir | ignore

    nu ($env.FILE_PWD | path join "extract.nu") --data-dir $dir
}

# Returns the first of the candidate globs (relative to root) that matches a file.
def find_pak [root: string, candidates: list<string>] {
    let hits = ($candidates | each {|c| glob ($root | path join $c)} | flatten)
    if ($hits | is-empty) {
        error make {msg: $"could not find ($candidates | str join ' or ') under ($root)"}
    }
    $hits | first
}
