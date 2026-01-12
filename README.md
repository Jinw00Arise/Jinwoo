# Jinwoo

A MapleStory v95 server emulator written in Go.

## Development Progress

### Core Infrastructure
- [x] Project structure setup (cmd/ and internal/)
- [x] Database connection and ORM
- [x] Crypto module (AES, Shanda encryption, IV handling)
- [x] Protocol packet system (read/write operations)
- [x] Network layer with connection handling
- [x] Graceful shutdown handling

### Login Server
- [x] Basic login server setup
- [x] Handshake and encryption
- [x] Opcode definitions (recv/send)
- [x] Packet handler routing
- [x] Account authentication
- [ ] World list display
- [ ] Character list display
- [ ] Character creation
- [ ] Character deletion
- [ ] Channel migration

### Channel Server
- [x] Basic channel server setup
- [ ] Player login and spawn
- [ ] Map handling
- [ ] Movement packets
- [ ] Chat system
- [ ] Inventory system
- [ ] Equipment handling
- [ ] NPC interaction
- [ ] Quest system
- [ ] Party system
- [ ] Trade system
- [ ] Mob spawning and AI
- [ ] Drop system
- [ ] Skill system
- [ ] Combat system

### Game Features
- [ ] Shops (NPC and player)
- [ ] Cash shop
- [ ] Storage system
- [ ] Buddy list
- [ ] Guild system
- [ ] Fame system
- [ ] Mini-games
- [ ] Events

### Database
- [x] Character repository
- [x] Inventory repository
- [ ] Account repository
- [ ] Item repository
- [ ] Map repository
- [ ] Mob repository
- [ ] NPC repository
- [ ] Quest repository

### Tools & Utilities
- [ ] Admin commands
- [ ] GM tools
- [ ] Server monitoring
- [ ] Data parsers (WZ files)
- [ ] Migration tools

## Getting Started
```bash
# Build login server
make build-login

# Build channel server
make build-channel

# Run servers
./bin/login
./bin/channel
```

## Configuration

Database connection and server settings are configured via environment variables or config files.