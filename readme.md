# Darkest Savior (WIP)

Darkest in Darkest Dungeon, and Savior in *Save Editor*, if you squint hard enough.

> Ruin has come to our command line.
>
> -- The Ancestor (probably)

## Installation

```shell
go get ...
```

## Usage

Open your command line:

```shell
# interactive mode
dsavior
# or
dsavior --interactive

# convert from/to DSON to JSON and vice versa
dsavior --convert --from path/to/persistent.dson.json --to path/to/result.json
dsavior --convert --from path/to/result.json --to path/to/persistent.dson.json
```

## Notes On DSON Files

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

- [x] Convert from DSON to JSON
- [ ] Convert from JSON to DSON
- [ ] Easier distribution for end user
- [ ] Interactive Mode
- [ ] GUI client
