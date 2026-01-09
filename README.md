# Jinwoo

A MapleStory v95 server emulator written in Go. This project aims to provide a clean, maintainable, and production-ready implementation of the MapleStory game server.

## Features

- **Login Server**: User authentication, world/channel selection, character creation
- **Channel Server**: Game world, NPCs, quests, inventory, drops, mobs, combat
- **Lua Scripting**: NPCs, portals, and quests are scripted using Lua
- **PostgreSQL Database**: Persistent storage for accounts, characters, and game data
- **WZ Data Loading**: XML-exported WZ files for game data (maps, items, mobs, quests)
- **Docker Support**: Ready for containerized deployment
- **Health Checks**: HTTP endpoints for monitoring and orchestration
- **Prometheus Metrics**: Performance and game metrics for monitoring

## Prerequisites

- Go 1.18+ (1.21+ recommended)
- PostgreSQL 13+
- MapleStory v95 client
- WZ files exported to XML format

## Quick Start

### 1. Clone and Build

```bash
git clone https://github.com/Jinw00Arise/Jinwoo.git
cd Jinwoo
make build
```

### 2. Configure Environment

Create a `.env` file in the project root:

```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/jinwoo?sslmode=disable

# Login Server
LOGIN_HOST=127.0.0.1
LOGIN_PORT=8484

# Channel Server
CHANNEL_HOST=127.0.0.1
CHANNEL_PORT=8585
WORLD_ID=0
CHANNEL_ID=0

# Game Configuration
WZ_PATH=./data/wz
SCRIPTS_PATH=./scripts
AUTO_REGISTER=1
EXP_RATE=1.0
QUEST_EXP_RATE=1.0

# Debug (optional)
DEBUG_PACKETS=0
```

### 3. Setup Database

```bash
# Create database
createdb jinwoo

# Run migrations
goose -dir migrations postgres "$DATABASE_URL" up
```

### 4. Prepare WZ Data

Export your WZ files to XML format and place them in `data/wz/`:

```
data/wz/
├── Map.wz/
├── Npc.wz/
├── Item.wz/
├── Mob.wz/
├── Quest.wz/
├── String.wz/
├── Base.wz/
└── Etc.wz/
```

### 5. Add Scripts

Place Lua scripts in the `scripts/` directory:

```
scripts/
├── npc/       # NPC conversation scripts
├── portal/    # Portal scripts
├── quest/     # Quest scripts
├── event/     # Event scripts (planned)
└── map/       # Map enter/exit scripts (planned)
```

### 6. Run the Server

```bash
make run
```

Or run servers separately:

```bash
make run-login    # Login server only
make run-channel  # Channel server only
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `LOGIN_HOST` | Login server bind address | `127.0.0.1` |
| `LOGIN_PORT` | Login server port | `8484` |
| `CHANNEL_HOST` | Channel server bind address | `127.0.0.1` |
| `CHANNEL_PORT` | Channel server port | `8585` |
| `WORLD_ID` | World identifier | `0` |
| `CHANNEL_ID` | Channel identifier | `0` |
| `WZ_PATH` | Path to WZ data directory | `./data/wz` |
| `SCRIPTS_PATH` | Path to Lua scripts | `./scripts` |
| `AUTO_REGISTER` | Auto-create accounts on login | `0` |
| `EXP_RATE` | Experience rate multiplier | `1.0` |
| `QUEST_EXP_RATE` | Quest experience rate multiplier | `1.0` |
| `DEBUG_PACKETS` | Enable packet debug logging | `0` |

## Architecture

```
Jinwoo/
├── cmd/
│   ├── login/          # Login server entry point
│   └── channel/        # Channel server entry point
├── internal/
│   ├── channel/        # Channel server implementation
│   │   ├── handler*.go # Packet handlers
│   │   ├── packets.go  # Packet builders
│   │   └── commands.go # GM commands
│   ├── crypto/         # AES/Shanda encryption
│   ├── database/       # GORM models and repositories
│   ├── game/           # Core game interfaces and logic
│   │   ├── field/      # Map/field implementation
│   │   ├── inventory/  # Inventory management
│   │   ├── session/    # Player session management
│   │   └── stage/      # Stage (map instance) system
│   ├── health/         # HTTP health check server
│   ├── login/          # Login server implementation
│   ├── network/        # TCP connection handling
│   ├── packet/         # Packet read/write utilities
│   ├── script/         # Lua scripting engine
│   └── wz/             # WZ data loader
├── pkg/maple/          # MapleStory constants and opcodes
├── scripts/            # Lua scripts
├── data/wz/            # WZ XML data files
├── migrations/         # Database migrations
└── config/             # Configuration loading
```

## GM Commands

GM commands are available in-game by typing in chat. Commands require GM level 1+.

| Command | Description | Usage |
|---------|-------------|-------|
| `!reloadscripts` | Hot-reload all Lua scripts | `!reloadscripts` |
| `!reloadwz` | Reload WZ data | `!reloadwz` |
| `!status` | Show server status | `!status` |
| `!map` | Warp to map | `!map <mapid>` |
| `!item` | Give item | `!item <itemid> [qty]` |
| `!level` | Set level | `!level <level>` |
| `!job` | Set job | `!job <jobid>` |
| `!heal` | Restore HP/MP | `!heal` |

## Scripting

### NPC Scripts

```lua
-- scripts/npc/2000.lua (Roger)
function start()
    say("Welcome to Maple Island!")
    
    if askYesNo("Would you like some help?") then
        say("Great! Let me show you around.")
    else
        say("Come back if you need anything!")
    end
end
```

### Portal Scripts

```lua
-- scripts/portal/glBmsg0.lua
function enter()
    balloonMessage("Once you leave this area you won't be able to return.", 150, 5)
end
```

### Quest Scripts

```lua
-- scripts/quest/1021s.lua (Roger's Apple - start)
function start()
    sayNext("Hey! I'm Roger, and I'll teach you the basics.")
    
    if not askAccept("Ready to begin?") then
        return
    end
    
    setHp(25)
    giveItem(2010007, 1) -- Roger's Apple
    forceStartQuest(1021)
    avatarOriented("UI/tutorial.img/28")
end
```

## Deployment

### Docker

```bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

### Manual

```bash
# Build
make build

# Run in production
./bin/jinwoo-login &
./bin/jinwoo-channel &
```

## Development

### Building

```bash
make build          # Build both servers
make build-login    # Build login server only
make build-channel  # Build channel server only
```

### Testing

```bash
make test           # Run all tests
go test -v ./...    # Verbose test output
```

### Code Quality

```bash
go vet ./...        # Run static analysis
go fmt ./...        # Format code
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
# {"status":"healthy","uptime":"1h30m45s"}

curl http://localhost:8080/ready
# ready
```

### Metrics (Prometheus)

```bash
curl http://localhost:8080/metrics
# jinwoo_connected_players 42
# jinwoo_active_stages 15
# jinwoo_uptime_seconds 5445
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please use [conventional commits](https://www.conventionalcommits.org/) for commit messages.

## License

This project is for educational purposes only. MapleStory is a trademark of Nexon.

## Acknowledgments

- [Kinoko](https://github.com/iw2d/kinoko) - Java v95 reference implementation
- [RustMS](https://github.com/jon-zu/reMember) - Rust MapleStory reference
- MapleStory community for protocol documentation
