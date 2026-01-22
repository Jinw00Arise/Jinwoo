# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Jinwoo is a MapleStory v95 server emulator written in Go. It's a unified game server handling both login and channel server functionality.

## Build & Run Commands

```bash
# Build the unified server
go build -o bin/jinwoo-server ./cmd/server

# Run the server (requires .env file and WZ data)
go run ./cmd/server

# Legacy separate builds (Makefile targets, may be outdated)
make build          # Build both login and channel
make build-login    # Build login server only
make build-channel  # Build channel server only
make clean          # Stop servers and remove binaries
```

## Configuration

Environment variables (`.env` file):
- `DATABASE_URL` - PostgreSQL connection string
- `LOGIN_HOST` / `LOGIN_PORT` - Login server bind address (default: 127.0.0.1:8484)
- `CHANNEL_HOST` / `CHANNEL_PORT` - Channel server bind address (default: 127.0.0.1:8585)
- `AUTO_REGISTER` - Enable automatic account registration (1=enabled)
- `WZ_PATH` - Path to WZ game data files (default: ./wz)
- `SCRIPTS_PATH` - Path to Lua scripts (default: scripts)

## Architecture

### Package Structure

```
cmd/server/          Entry point - unified server
internal/
├── game/
│   ├── server/      Core server (handlers, config, client management)
│   ├── field/       Game world entities (Character, NPC, Mob, Field)
│   ├── packets/     Packet definitions and opcodes
│   ├── quest/       Quest system
│   └── script/      Lua scripting (NPC/portal scripts)
├── data/
│   ├── providers/   WZ data providers (Items, Maps, Quests, NPCs)
│   ├── repositories/ Database access (Accounts, Characters, Items)
│   └── db/          Database connection
├── database/models/ GORM models (Account, Character, CharacterItem, etc.)
├── protocol/        Packet read/write, binary encoding
├── crypto/          AES encryption, Shanda cipher, IV handling
└── network/         TCP connection wrapper
```

### Key Concepts

**Server Flow**: Client connects → Handshake (encryption setup) → Login authentication → World/channel selection → Character migration to channel → In-game packets

**Field System**: Maps are loaded lazily from WZ data. Each Field contains Characters, NPCs, and Mobs managed by their respective managers. Entities broadcast packets to each other.

**Packet System**: Opcodes defined in `internal/game/packets/opcodes.go` (channel) and `internal/game/server/opcodes.go` (login). Packets use little-endian encoding via `protocol.Packet`.

**Scripting**: Lua scripts in `scripts/npc/` and `scripts/portal/` are executed via gopher-lua. Scripts interact with the game through `CharacterAccessor` interface providing player data and actions.

**Data Providers**: WZ files (MapleStory game data format) are parsed by `internal/data/providers/wz/` and exposed through typed providers (ItemProvider, MapProvider, etc.).

### Database

PostgreSQL with GORM. Models auto-migrate on startup. Key tables: accounts, characters, character_items, quest_records.

### Encryption

MapleStory v95 protocol uses:
- AES for packet encryption
- Shanda cipher as additional layer
- Rolling IV (initialization vector) per connection

## Working with Opcodes

Opcodes vary by MapleStory client version. When dialogs/packets don't work, verify the opcode matches the client:
- Check `internal/game/packets/opcodes.go` for channel opcodes
- Check `internal/game/server/opcodes.go` for login opcodes
- Common v95 opcodes may differ from v83 clients

**IMPORTANT**: Always ask the user before adding or changing an opcode value. Do not guess opcode values - they must be verified against the specific client being used.
