# Darkest Savior

Darkest in Darkest Dungeon, and Savior in *Save Editor*, if you squint hard enough.

> Ruin has come to our command line.
>
> -- The Ancestor (probably)

## Installation

```shell
go install github.com/thanhnguyen2187/darkest-savior@master

darkest-savior --help
# Ruin has come to our command line.
# 
# A CLI utility to convert DSON (Darkest Dungeon's own proprietary JSON format)
# to "standard" JSON in the command line.
# 
# Usage: main <command> [<args>]
# 
# Options:
#   --help, -h             display this help and exit
# 
# Commands:
#   convert
```

## Usage

Open your command line:

```shell
# convert from DSON to JSON
darkest-savior convert \
    --from sample_dson/persistent.campaign_log.json \
    --to sample_json/persistent.campaign_log.json

# for safety reason, --force is needed if the result file existed
darkest-savior convert \
    --force \
    --from sample_dson/persistent.campaign_log.json \
    --to sample_json/persistent.campaign_log.json

# convert from JSON to DSON
darkest-savior convert \
    --from sample_json/persistent.campaign_log.json \
    --to sample_dson/persistent.campaign_log.json
```

## Notes On DSON Files

You can have a look at the converted files yourself in folder `sample_json`.

- `novelty_tracker.json`: in-game elements that the player encountered (building, trinkets, etc.)
- `persist.campaign_log.json`: expeditions (dungeon runs) history
- `persist.campaign_mash.json`: unknown
- `persist.curio_tracker.json`: unknown
- `persist.estate.json`: trinkets and resources (gold, bust, portrait, etc.)
- `persist.game.json`: save file options (name, enabled DLCs, etc.)
- `persist.game_knowledge.json`: unlocked dungeons, and monster skills
- `persist.journal.json`: unlocked journal pages?
- `persist.narration.json`: narrator voice lines settings
- `persist.progression.json`: quest completing progression
- `persist.quest.json`: current open quests
- `persist.roster.json`: current heroes and quirks
- `persist.town_event.json`: current town event
- `persist.town.json`: town upgrades and state (which hero is being treated at Sanitarium, etc.)
- `persist.tutorial.json`: in-game elements (that have help suggestion) that the player encountered
- `persist.upgrades.json`: building and heroes upgrade history

## TODO

- [x] Convert from DSON to JSON: done
- [x] Convert from JSON to DSON: done
- [ ] Easier distribution for end user: using `go install ...` is not the best way to distribute the tool, so the plan
  is to build binary files for different platforms
- [ ] Interactive Mode: a TUI client
- [ ] GUI Client
