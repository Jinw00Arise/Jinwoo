# Scripts Directory

Lua scripts for NPCs, portals, quests, and events.

## Directory Structure

```
scripts/
├── npc/        # NPC dialogue scripts (named by NPC ID)
├── portal/     # Portal scripts
├── quest/      # Quest scripts
├── event/      # Event scripts (GM events, etc.)
└── map/        # Map enter/exit scripts
```

## NPC Script Example

```lua
-- scripts/npc/9010000.lua (Maple Administrator)

function start()
    -- Simple dialogue
    say("Hello, #h#!")  -- #h# = player name
    
    -- Yes/No question
    local answer = askYesNo("Do you want some items?")
    if answer then
        giveItem(2000000, 10)  -- 10 Red Potions
        say("Here you go!")
    end
end
```

## Portal Script Example

```lua
-- scripts/portal/10000_glBmsg0.lua
-- Shows warning when leaving Mushroom Town

balloonMessage("Once you leave this area you won't be able to return.", 150, 5)
```

Portal scripts can use:
- `balloonMessage(text, width, duration)` - Show balloon above player
- `warp(mapId, portalName)` - Teleport player to map
- `getLevel()` - Get player level
- `getJob()` - Get player job
- `getMapId()` - Get current map
- `log(msg)` - Log to server console

## Available Functions (NPC Scripts)

### Player Info
- `getPlayerName()` - Returns player's name
- `getPlayerLevel()` - Returns player's level
- `getPlayerJob()` - Returns player's job ID
- `getPlayerMeso()` - Returns player's meso
- `getPlayerMap()` - Returns current map ID

### Dialogue
- `say(text)` - Show message with OK button
- `sayNext(text)` - Show message with Next button
- `sayPrev(text)` - Show message with Back/Next buttons
- `askYesNo(text)` - Yes/No question, returns true/false
- `askMenu(text)` - Selection menu, returns selection index (0-based)
- `askNumber(text, default, min, max)` - Number input
- `askText(text, default, minLen, maxLen)` - Text input
- `askAcceptDecline(text)` - Accept/Decline question

### Actions
- `warp(mapId, portal)` - Teleport player
- `giveExp(amount)` - Give EXP
- `giveMeso(amount)` - Give meso
- `giveItem(itemId, count)` - Give item, returns success
- `hasItem(itemId, count)` - Check if player has item
- `takeItem(itemId, count)` - Remove item from player
- `gainFame(amount)` - Give fame

### Utility
- `endChat()` - End conversation immediately
- `log(message)` - Log message to server console

## Text Codes

- `#h#` or `#H#` - Player name
- `#e` - Bold text start
- `#n` - Normal text (end bold)
- `#b` - Blue color
- `#r` - Red color
- `#k` - Black color (default)
- `#L0#text#l` - Menu option (0 = first option)
- `#v1234567#` - Show item icon
- `#t1234567#` - Show item name
- `#c1234567#` - Show item count in inventory
- `\r\n` - New line

## Hot Reloading

Scripts are loaded at server start. To reload without restart:
1. Edit the script file
2. Use a GM command to reload scripts (TODO)

## File Naming

- NPC scripts: `<npc_id>.lua` (e.g., `9010000.lua`)
- Portal scripts: `<portal_name>.lua` or `<map_id>_<portal_name>.lua`
- Quest scripts: `<quest_id>.lua`

