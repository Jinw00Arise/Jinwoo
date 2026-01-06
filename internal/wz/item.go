package wz

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

// ItemData contains parsed item information
type ItemData struct {
	ID        int32
	Name      string
	Desc      string
	SlotMax   int16                // Max stack size (0 = not stackable/equip)
	InvType   models.InventoryType // Which inventory tab
	Price     int32
	Quest     bool // Quest item (cannot be traded)
	Cash      bool // Cash shop item
	TradeBlock bool
	
	// Consumable effects (from spec node)
	HP        int32 // Flat HP recovery
	MP        int32 // Flat MP recovery
	HPR       int32 // Percentage HP recovery (0-100)
	MPR       int32 // Percentage MP recovery (0-100)
	Time      int32 // Buff duration in seconds
	Speed     int32 // Speed buff
	Jump      int32 // Jump buff
	Attack    int32 // Attack buff
	MAttack   int32 // Magic attack buff
	Defense   int32 // Defense buff
	MDefense  int32 // Magic defense buff
	Accuracy  int32 // Accuracy buff
	Avoidability int32 // Avoidability buff
	MoveTo    int32 // Teleport map ID (for return scrolls)
	Cure      bool  // Cures status effects
}

// Default stack sizes by inventory type
const (
	DefaultConsumeStack = 100
	DefaultEtcStack     = 100
	DefaultInstallStack = 100
	DefaultCashStack    = 1
	DefaultEquipStack   = 1
)

// GetInventoryTypeFromItemID determines inventory type from item ID
func GetInventoryTypeFromItemID(itemID int32) models.InventoryType {
	prefix := itemID / 1000000
	switch prefix {
	case 1: // 1xxxxxx = Equipment
		return models.InventoryEquip
	case 2: // 2xxxxxx = Consumables
		return models.InventoryConsume
	case 3: // 3xxxxxx = Setup/Install
		return models.InventoryInstall
	case 4: // 4xxxxxx = Etc
		return models.InventoryEtc
	case 5: // 5xxxxxx = Cash
		return models.InventoryCash
	default:
		return models.InventoryEtc
	}
}

// IsEquipItem returns true if the item is an equipment
func IsEquipItem(itemID int32) bool {
	return itemID/1000000 == 1
}

// GetDefaultSlotMax returns the default stack size for an item type
func GetDefaultSlotMax(itemID int32) int16 {
	if IsEquipItem(itemID) {
		return DefaultEquipStack
	}
	invType := GetInventoryTypeFromItemID(itemID)
	switch invType {
	case models.InventoryConsume:
		return DefaultConsumeStack
	case models.InventoryInstall:
		return DefaultInstallStack
	case models.InventoryEtc:
		return DefaultEtcStack
	case models.InventoryCash:
		return DefaultCashStack
	default:
		return 1
	}
}

// LoadItemData loads item information from WZ files
func (dm *DataManager) LoadItemData(itemID int32) *ItemData {
	dm.mu.RLock()
	if item, exists := dm.items[itemID]; exists {
		dm.mu.RUnlock()
		return item
	}
	dm.mu.RUnlock()

	item := dm.parseItemData(itemID)
	if item != nil {
		dm.mu.Lock()
		dm.items[itemID] = item
		dm.mu.Unlock()
	}
	return item
}

func (dm *DataManager) parseItemData(itemID int32) *ItemData {
	item := &ItemData{
		ID:      itemID,
		InvType: GetInventoryTypeFromItemID(itemID),
		SlotMax: GetDefaultSlotMax(itemID),
	}

	// Determine the WZ path based on item type
	var itemPath string
	prefix := itemID / 1000000
	subPrefix := (itemID / 10000) % 100

	switch prefix {
	case 1: // Equipment - handled differently (Character.wz)
		item.SlotMax = 1
		return item
	case 2: // Consumables
		itemPath = filepath.Join(dm.basePath, "Item.wz", "Consume", fmt.Sprintf("%04d.img.xml", itemID/10000))
	case 3: // Setup/Install
		itemPath = filepath.Join(dm.basePath, "Item.wz", "Install", fmt.Sprintf("%04d.img.xml", itemID/10000))
	case 4: // Etc
		itemPath = filepath.Join(dm.basePath, "Item.wz", "Etc", fmt.Sprintf("%04d.img.xml", itemID/10000))
	case 5: // Cash
		if subPrefix >= 0 && subPrefix <= 9 {
			itemPath = filepath.Join(dm.basePath, "Item.wz", "Cash", fmt.Sprintf("%04d.img.xml", itemID/10000))
		} else {
			itemPath = filepath.Join(dm.basePath, "Item.wz", "Pet", fmt.Sprintf("%07d.img.xml", itemID))
		}
	default:
		return item
	}

	// Parse the item XML file
	root, err := ParseFile(itemPath)
	if err != nil {
		// File might not exist for some items
		log.Printf("[WZ] Failed to load item file %s: %v", itemPath, err)
		return item
	}
	log.Printf("[WZ] Loaded item file %s", itemPath)

	// Find the specific item node
	itemIDStr := fmt.Sprintf("%08d", itemID)
	itemNode := root.GetChild(itemIDStr)
	if itemNode == nil {
		// Try without leading zeros
		itemIDStr = strconv.Itoa(int(itemID))
		itemNode = root.GetChild(itemIDStr)
	}
	if itemNode == nil {
		log.Printf("[WZ] Item node %d not found in %s (tried %08d and %d)", itemID, itemPath, itemID, itemID)
		return item
	}
	log.Printf("[WZ] Found item node for %d", itemID)

	// Parse info node
	info := itemNode.GetChild("info")
	if info != nil {
		if slotMax := info.GetInt("slotMax"); slotMax > 0 {
			item.SlotMax = int16(slotMax)
		}
		item.Price = int32(info.GetInt("price"))
		item.Quest = info.GetInt("quest") == 1
		item.Cash = info.GetInt("cash") == 1
		item.TradeBlock = info.GetInt("tradeBlock") == 1
	}
	
	// Parse spec node (consumable effects)
	spec := itemNode.GetChild("spec")
	if spec != nil {
		item.HP = int32(spec.GetInt("hp"))
		item.MP = int32(spec.GetInt("mp"))
		item.HPR = int32(spec.GetInt("hpR"))
		item.MPR = int32(spec.GetInt("mpR"))
		item.Time = int32(spec.GetInt("time"))
		item.Speed = int32(spec.GetInt("speed"))
		item.Jump = int32(spec.GetInt("jump"))
		item.Attack = int32(spec.GetInt("pad")) // Physical Attack Damage
		item.MAttack = int32(spec.GetInt("mad")) // Magic Attack Damage
		item.Defense = int32(spec.GetInt("pdd")) // Physical Defense
		item.MDefense = int32(spec.GetInt("mdd")) // Magic Defense
		item.Accuracy = int32(spec.GetInt("acc"))
		item.Avoidability = int32(spec.GetInt("eva"))
		item.MoveTo = int32(spec.GetInt("moveTo"))
		item.Cure = spec.GetInt("cure") == 1
		log.Printf("[WZ] Item %d spec: hp=%d, mp=%d, hpR=%d, mpR=%d", itemID, item.HP, item.MP, item.HPR, item.MPR)
	} else {
		log.Printf("[WZ] Item %d has no spec node", itemID)
	}

	// Load item name from String.wz
	item.Name, item.Desc = dm.loadItemString(itemID)

	return item
}

func (dm *DataManager) loadItemString(itemID int32) (name, desc string) {
	prefix := itemID / 1000000
	var stringPath string
	var category string

	switch prefix {
	case 1: // Equipment
		stringPath = filepath.Join(dm.basePath, "String.wz", "Eqp.img.xml")
		category = "Eqp"
	case 2: // Consumables
		stringPath = filepath.Join(dm.basePath, "String.wz", "Consume.img.xml")
		category = "Con"
	case 3: // Install
		stringPath = filepath.Join(dm.basePath, "String.wz", "Ins.img.xml")
		category = "Ins"
	case 4: // Etc
		stringPath = filepath.Join(dm.basePath, "String.wz", "Etc.img.xml")
		category = "Etc"
	case 5: // Cash/Pet
		stringPath = filepath.Join(dm.basePath, "String.wz", "Cash.img.xml")
		category = "Cash"
	default:
		return "", ""
	}

	root, err := ParseFile(stringPath)
	if err != nil {
		return "", ""
	}

	// Different structure for Eqp.img.xml
	if category == "Eqp" {
		return dm.loadEqpItemString(root, itemID)
	}

	// Standard structure: direct children with item IDs
	itemIDStr := strconv.Itoa(int(itemID))
	itemNode := root.GetChild(itemIDStr)
	if itemNode == nil {
		return "", ""
	}

	return itemNode.GetString("name"), itemNode.GetString("desc")
}

func (dm *DataManager) loadEqpItemString(root *Node, itemID int32) (name, desc string) {
	// Eqp.img.xml has nested structure: Eqp -> Category -> Items
	eqpNode := root.GetChild("Eqp")
	if eqpNode == nil {
		return "", ""
	}

	itemIDStr := strconv.Itoa(int(itemID))

	// Search through all equipment categories
	for _, categoryNode := range eqpNode.GetAllChildren() {
		itemNode := categoryNode.GetChild(itemIDStr)
		if itemNode != nil {
			return itemNode.GetString("name"), itemNode.GetString("desc")
		}
	}

	return "", ""
}

// GetItemName returns just the item name
func (dm *DataManager) GetItemName(itemID int32) string {
	item := dm.LoadItemData(itemID)
	if item != nil && item.Name != "" {
		return item.Name
	}
	return fmt.Sprintf("Unknown Item (%d)", itemID)
}

// GetSlotMax returns the max stack size for an item
func (dm *DataManager) GetSlotMax(itemID int32) int16 {
	item := dm.LoadItemData(itemID)
	if item != nil {
		return item.SlotMax
	}
	return GetDefaultSlotMax(itemID)
}

// GetItemRecovery returns HP and MP recovery amounts for a consumable item
// Returns flat HP recovery, flat MP recovery, HP percentage, MP percentage
func (dm *DataManager) GetItemRecovery(itemID int32) (hp, mp, hpR, mpR int32) {
	item := dm.LoadItemData(itemID)
	if item != nil {
		return item.HP, item.MP, item.HPR, item.MPR
	}
	return 0, 0, 0, 0
}

// GetItemEffects returns all consumable effects for an item
func (dm *DataManager) GetItemEffects(itemID int32) *ItemData {
	return dm.LoadItemData(itemID)
}

// LoadAllItemStrings preloads item string data for quick name lookups
// NOTE: This only caches names/descriptions, not spec data.
// Full item data (including spec) is loaded on-demand by LoadItemData.
func (dm *DataManager) LoadAllItemStrings() {
	// We no longer pre-cache items here to avoid caching incomplete data.
	// Item names are loaded as part of LoadItemData -> loadItemString.
	// This function is kept for backward compatibility but now just logs.
	log.Printf("[WZ] Item string loading deferred to on-demand")
}

// CanStack returns true if items can be stacked together
func CanStack(item1, item2 *models.Inventory) bool {
	// Items must be same item ID
	if item1.ItemID != item2.ItemID {
		return false
	}
	// Equipment cannot stack
	if item1.IsEquip() {
		return false
	}
	return true
}

// ParseEquipStats parses equipment bonus stats from WZ data
func (dm *DataManager) ParseEquipStats(itemID int32) map[string]int16 {
	stats := make(map[string]int16)
	
	// Equipment is in Character.wz, not Item.wz
	// For now, return empty stats - can be extended later
	// Most equip creation happens through character creation anyway
	
	return stats
}

// Helper to get item file prefix string
func getItemFilePrefix(itemID int32) string {
	return fmt.Sprintf("%04d", itemID/10000)
}

// ValidateItemID checks if an item ID is valid
func ValidateItemID(itemID int32) bool {
	prefix := itemID / 1000000
	return prefix >= 1 && prefix <= 5
}

// GetItemCategory returns a human-readable category name
func GetItemCategory(itemID int32) string {
	prefix := itemID / 1000000
	switch prefix {
	case 1:
		subPrefix := (itemID / 10000) % 100
		categories := map[int]string{
			0: "Cap", 1: "Face Accessory", 2: "Eye Accessory", 3: "Earring",
			4: "Topwear", 5: "Overall", 6: "Bottomwear", 7: "Shoes",
			8: "Gloves", 9: "Shield", 10: "Cape", 11: "Ring",
			12: "Pendant", 13: "Belt", 14: "Medal",
			30: "One-Handed Sword", 31: "One-Handed Axe", 32: "One-Handed BW",
			33: "Dagger", 34: "Katara", 37: "Wand", 38: "Staff",
			40: "Two-Handed Sword", 41: "Two-Handed Axe", 42: "Two-Handed BW",
			43: "Spear", 44: "Polearm", 45: "Bow", 46: "Crossbow",
			47: "Claw", 48: "Knuckle", 49: "Gun",
		}
		if cat, ok := categories[int(subPrefix)]; ok {
			return cat
		}
		return "Equipment"
	case 2:
		return "Consumable"
	case 3:
		return "Setup"
	case 4:
		return "Etc"
	case 5:
		return "Cash"
	default:
		return "Unknown"
	}
}

// IsCashItem returns true if the item is from Cash Shop
func IsCashItem(itemID int32) bool {
	return itemID/1000000 == 5
}

// IsQuestItem checks if item is a quest item based on WZ data
func (dm *DataManager) IsQuestItem(itemID int32) bool {
	item := dm.LoadItemData(itemID)
	return item != nil && item.Quest
}

// SplitItemID into prefix and suffix for debugging
func SplitItemID(itemID int32) (prefix int32, suffix int32) {
	return itemID / 10000, itemID % 10000
}

// Helper to check if string parsing is needed
func needsStringParsing(s string) bool {
	return strings.Contains(s, "#") || strings.Contains(s, "\\n")
}

