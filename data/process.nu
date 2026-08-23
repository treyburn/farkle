#!/usr/bin/env nu

# Rebuild dice.json from the KCD2 game xml files.
#
# Expects item.xml, item__dlc.xml and text_ui_items.xml to already be present in
# the data directory - `sync.nu` pulls those out of a game install for you.
def main [
    --data-dir: string # Directory holding the xml files and the dice.json output. Defaults to this script's directory.
] {
    let dir = ($data_dir | default $env.FILE_PWD)

    let item_xml = ($dir | path join "item.xml")
    let dlc_xml = ($dir | path join "item__dlc.xml")
    let ui_xml = ($dir | path join "text_ui_items.xml")
    let out = ($dir | path join "dice.json")

    for file in [$item_xml $dlc_xml $ui_xml] {
        if not ($file | path exists) {
            error make {msg: $"missing ($file) - run `just data-sync <game-dir>` to extract it from a game install"}
        }
    }

    let dice = (
        open $item_xml
        | get content.0.content
        | where tag == 'Die'
        | get attributes
    )

    let dlc_dice = (
        open $dlc_xml
        | get content.0.content
        | where tag == 'Die'
        | get attributes
    )

    # Quest item flag is useful if we need to de-dupe identical dice - like in the case with Lucky Die.
    let all_dice = ($dice | append $dlc_dice | default "false" IsQuestItem)
    let die_names = ($all_dice | get UIName)

    let ui_map = (
        open $ui_xml
        | get content
        | get content
        | where {|row| $row.0.content.0.content in $die_names}
        | each {|row| {
            UIName: $row.0.content.0.content
            DisplayName: $row.2.content.0.content
          }}
    )

    let result = (
        $all_dice
        | join $ui_map UIName
        # There are 2 instances of LuckyDie in the game. They are identical in states but one is a quest item and the other is not.
        # Instead of having those duplicate - I choose to collapse them into a single result.
        | group-by {|r| $"($r.DisplayName)|($r.Price)|($r.SideWeights | str trim)|($r.SideValues | str trim)"}
        | values
        | each {|g| $g | sort-by IsQuestItem | first}
        | select Id SideWeights SideValues DisplayName
    )

    $result | save -f $out

    print $"wrote ($result | length) dice to ($out)"
}
