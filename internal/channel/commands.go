package channel

import (
	"fmt"
	"log"
	"strings"

	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/game/stage"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// GMLevel constants
const (
	GMPlayerLevel = 0 // Regular player
	GMLevel       = 1 // GM commands
	AdminLevel    = 2 // Admin commands
)

// CommandHandler handles GM/Admin commands from chat
type CommandHandler struct {
	stageManager *stage.StageManager
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(sm *stage.StageManager) *CommandHandler {
	return &CommandHandler{
		stageManager: sm,
	}
}

// ProcessCommand processes a GM command and returns (response message, handled)
func (c *CommandHandler) ProcessCommand(user *stage.User, gmLevel int, message string) (string, bool) {
	if !strings.HasPrefix(message, "!") {
		return "", false
	}

	parts := strings.Fields(message)
	if len(parts) == 0 {
		return "", false
	}

	cmd := strings.ToLower(parts[0][1:]) // Remove ! prefix
	args := parts[1:]

	switch cmd {
	case "help":
		return c.cmdHelp(gmLevel), true

	case "reloadscripts":
		if gmLevel < AdminLevel {
			return "Insufficient permissions", true
		}
		return c.cmdReloadScripts(), true

	case "reloadwz":
		if gmLevel < AdminLevel {
			return "Insufficient permissions", true
		}
		return c.cmdReloadWZ(), true

	case "status":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdStatus(), true

	case "pos":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdPos(user), true

	case "map":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdMap(user, args), true

	case "item":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdItem(user, args), true

	case "level":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdLevel(user, args), true

	case "job":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdJob(user, args), true

	case "heal":
		if gmLevel < GMLevel {
			return "Insufficient permissions", true
		}
		return c.cmdHeal(user), true

	default:
		return fmt.Sprintf("Unknown command: %s. Type !help for available commands.", cmd), true
	}
}

func (c *CommandHandler) cmdHelp(gmLevel int) string {
	help := "Available commands:\n"
	help += "!help - Show this help\n"
	
	if gmLevel >= GMLevel {
		help += "!pos - Show current position\n"
		help += "!map <id> - Warp to map\n"
		help += "!item <id> [qty] - Give item\n"
		help += "!level <level> - Set level\n"
		help += "!job <id> - Set job\n"
		help += "!heal - Restore HP/MP\n"
		help += "!status - Server status\n"
	}
	
	if gmLevel >= AdminLevel {
		help += "!reloadscripts - Reload Lua scripts\n"
		help += "!reloadwz - Reload WZ data\n"
	}
	
	return help
}

func (c *CommandHandler) cmdReloadScripts() string {
	manager := script.GetManager()
	if manager == nil {
		return "Script manager not available"
	}

	if err := manager.ReloadScripts(); err != nil {
		log.Printf("[Command] Failed to reload scripts: %v", err)
		return fmt.Sprintf("Failed to reload scripts: %v", err)
	}

	log.Printf("[Command] Scripts reloaded successfully")
	return "Scripts reloaded successfully"
}

func (c *CommandHandler) cmdReloadWZ() string {
	dm := wz.GetInstance()
	if dm == nil {
		return "WZ data manager not available"
	}

	if err := dm.Reload(); err != nil {
		log.Printf("[Command] Failed to reload WZ data: %v", err)
		return fmt.Sprintf("Failed to reload WZ data: %v", err)
	}

	log.Printf("[Command] WZ data reloaded successfully")
	return "WZ data reloaded successfully"
}

func (c *CommandHandler) cmdStatus() string {
	if c.stageManager == nil {
		return "Stage manager not available"
	}

	stages := c.stageManager.Count()
	users := c.stageManager.TotalUsers()

	return fmt.Sprintf("Server Status:\n- Active stages: %d\n- Connected players: %d", stages, users)
}

func (c *CommandHandler) cmdPos(user *stage.User) string {
	x, y := user.Position()
	mapID := user.MapID()
	return fmt.Sprintf("Position: Map=%d, X=%d, Y=%d", mapID, x, y)
}

func (c *CommandHandler) cmdMap(user *stage.User, args []string) string {
	if len(args) < 1 {
		return "Usage: !map <mapid>"
	}

	var mapID int32
	_, err := fmt.Sscanf(args[0], "%d", &mapID)
	if err != nil {
		return "Invalid map ID"
	}

	// TODO: Implement map warp through stage manager
	return fmt.Sprintf("Map warp to %d not yet implemented", mapID)
}

func (c *CommandHandler) cmdItem(user *stage.User, args []string) string {
	if len(args) < 1 {
		return "Usage: !item <itemid> [quantity]"
	}

	var itemID int32
	quantity := int16(1)

	_, err := fmt.Sscanf(args[0], "%d", &itemID)
	if err != nil {
		return "Invalid item ID"
	}

	if len(args) >= 2 {
		var qty int
		_, err2 := fmt.Sscanf(args[1], "%d", &qty)
		if err2 == nil && qty > 0 {
			quantity = int16(qty)
		}
	}

	inv := user.Inventory()
	if inv == nil {
		return "Inventory not available"
	}

	ops, err := inv.AddItem(itemID, quantity)
	if err != nil {
		return fmt.Sprintf("Failed to add item: %v", err)
	}

	// Send inventory update to client
	if len(ops) > 0 {
		_ = user.Write(inventory.InventoryOperationPacket(ops, true)) // Ignore send errors
	}

	// Get item name
	dm := wz.GetInstance()
	itemName := dm.GetItemName(itemID)
	if itemName == "" {
		itemName = fmt.Sprintf("%d", itemID)
	}

	return fmt.Sprintf("Gave %d x %s", quantity, itemName)
}

func (c *CommandHandler) cmdLevel(user *stage.User, args []string) string {
	if len(args) < 1 {
		return "Usage: !level <level>"
	}

	var level int
	_, err := fmt.Sscanf(args[0], "%d", &level)
	if err != nil || level < 1 || level > 200 {
		return "Invalid level (1-200)"
	}

	char := user.Character()
	if char == nil {
		return "Character not available"
	}

	char.Level = byte(level)
	char.EXP = 0

	// Send stat update
	stats := map[uint32]int64{
		StatLevel: int64(char.Level),
		StatEXP:   int64(char.EXP),
	}
	user.Write(StatChangedPacket(false, stats))

	return fmt.Sprintf("Level set to %d", level)
}

func (c *CommandHandler) cmdJob(user *stage.User, args []string) string {
	if len(args) < 1 {
		return "Usage: !job <jobid>"
	}

	var jobID int16
	_, err := fmt.Sscanf(args[0], "%d", &jobID)
	if err != nil {
		return "Invalid job ID"
	}

	char := user.Character()
	if char == nil {
		return "Character not available"
	}

	char.Job = jobID

	// Send stat update
	stats := map[uint32]int64{StatJob: int64(char.Job)}
	user.Write(StatChangedPacket(false, stats))

	return fmt.Sprintf("Job set to %d", jobID)
}

func (c *CommandHandler) cmdHeal(user *stage.User) string {
	char := user.Character()
	if char == nil {
		return "Character not available"
	}

	char.HP = char.MaxHP
	char.MP = char.MaxMP

	// Send stat update
	stats := map[uint32]int64{
		StatHP: int64(char.HP),
		StatMP: int64(char.MP),
	}
	user.Write(StatChangedPacket(false, stats))

	return "HP and MP restored"
}

