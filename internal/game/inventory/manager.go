package inventory

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// ctx returns a background context for database operations
// TODO: Propagate context from handler to inventory operations
func ctx() context.Context {
	return context.Background()
}

// InventorySlots defines the maximum slots per inventory type
const (
	EquipSlots   = 24
	ConsumeSlots = 24
	InstallSlots = 24
	EtcSlots     = 24
	CashSlots    = 24
)

// Manager handles inventory operations for a character
type Manager struct {
	characterID uint
	repo        *repository.InventoryRepository
	
	// In-memory cache of inventory items
	equipped  map[int16]*models.Inventory // slot -> item (negative slots)
	equip     map[int16]*models.Inventory // slot -> item (1-24)
	consume   map[int16]*models.Inventory // slot -> item (1-24)
	install   map[int16]*models.Inventory // slot -> item (1-24)
	etc       map[int16]*models.Inventory // slot -> item (1-24)
	cash      map[int16]*models.Inventory // slot -> item (1-24)
	
	// Slot capacities (can be expanded)
	equipCapacity   byte
	consumeCapacity byte
	installCapacity byte
	etcCapacity     byte
	cashCapacity    byte
	
	mu sync.RWMutex
}

// NewManager creates a new inventory manager for a character
func NewManager(characterID uint, repo *repository.InventoryRepository) *Manager {
	return &Manager{
		characterID:     characterID,
		repo:            repo,
		equipped:        make(map[int16]*models.Inventory),
		equip:           make(map[int16]*models.Inventory),
		consume:         make(map[int16]*models.Inventory),
		install:         make(map[int16]*models.Inventory),
		etc:             make(map[int16]*models.Inventory),
		cash:            make(map[int16]*models.Inventory),
		equipCapacity:   EquipSlots,
		consumeCapacity: ConsumeSlots,
		installCapacity: InstallSlots,
		etcCapacity:     EtcSlots,
		cashCapacity:    CashSlots,
	}
}

// Load loads all inventory items from the database
func (m *Manager) Load() error {
	items, err := m.repo.FindByCharacterID(ctx(), m.characterID)
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, item := range items {
		m.cacheItem(item)
	}
	
	log.Printf("[Inventory] Loaded %d items for character %d", len(items), m.characterID)
	return nil
}

// cacheItem adds an item to the appropriate in-memory cache
func (m *Manager) cacheItem(item *models.Inventory) {
	invType := models.InventoryType(item.Type)
	
	switch invType {
	case models.InventoryEquipped:
		m.equipped[item.Slot] = item
	case models.InventoryEquip:
		m.equip[item.Slot] = item
	case models.InventoryConsume:
		m.consume[item.Slot] = item
	case models.InventoryInstall:
		m.install[item.Slot] = item
	case models.InventoryEtc:
		m.etc[item.Slot] = item
	case models.InventoryCash:
		m.cash[item.Slot] = item
	}
}

// removeFromCache removes an item from the in-memory cache
func (m *Manager) removeFromCache(item *models.Inventory) {
	invType := models.InventoryType(item.Type)
	
	switch invType {
	case models.InventoryEquipped:
		delete(m.equipped, item.Slot)
	case models.InventoryEquip:
		delete(m.equip, item.Slot)
	case models.InventoryConsume:
		delete(m.consume, item.Slot)
	case models.InventoryInstall:
		delete(m.install, item.Slot)
	case models.InventoryEtc:
		delete(m.etc, item.Slot)
	case models.InventoryCash:
		delete(m.cash, item.Slot)
	}
}

// getInventoryMap returns the appropriate inventory map for the given type
func (m *Manager) getInventoryMap(invType models.InventoryType) map[int16]*models.Inventory {
	switch invType {
	case models.InventoryEquipped:
		return m.equipped
	case models.InventoryEquip:
		return m.equip
	case models.InventoryConsume:
		return m.consume
	case models.InventoryInstall:
		return m.install
	case models.InventoryEtc:
		return m.etc
	case models.InventoryCash:
		return m.cash
	default:
		return nil
	}
}

// GetCapacity returns the capacity for an inventory type
func (m *Manager) GetCapacity(invType models.InventoryType) byte {
	switch invType {
	case models.InventoryEquip:
		return m.equipCapacity
	case models.InventoryConsume:
		return m.consumeCapacity
	case models.InventoryInstall:
		return m.installCapacity
	case models.InventoryEtc:
		return m.etcCapacity
	case models.InventoryCash:
		return m.cashCapacity
	default:
		return 0
	}
}

// findFreeSlot finds the first free slot in an inventory
func (m *Manager) findFreeSlot(invType models.InventoryType) int16 {
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return 0
	}
	
	capacity := m.GetCapacity(invType)
	for slot := int16(1); slot <= int16(capacity); slot++ {
		if _, exists := invMap[slot]; !exists {
			return slot
		}
	}
	return 0 // No free slot
}

// findExistingStack finds an existing stack of items that can accept more
func (m *Manager) findExistingStack(itemID int32, invType models.InventoryType) *models.Inventory {
	dm := wz.GetInstance()
	maxStack := dm.GetSlotMax(itemID)
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil
	}
	
	for _, item := range invMap {
		if item.ItemID == itemID && item.Quantity < maxStack {
			return item
		}
	}
	return nil
}

// AddItem adds an item to the inventory
// Returns the operations performed (for sending to client)
func (m *Manager) AddItem(itemID int32, quantity int16) ([]*InventoryOperation, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity: %d", quantity)
	}
	
	dm := wz.GetInstance()
	invType := wz.GetInventoryTypeFromItemID(itemID)
	maxStack := dm.GetSlotMax(itemID)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var operations []*InventoryOperation
	remaining := quantity
	
	// For equipment, always create new slot
	if wz.IsEquipItem(itemID) {
		slot := m.findFreeSlot(invType)
		if slot == 0 {
			return nil, fmt.Errorf("inventory full")
		}
		
		item := &models.Inventory{
			CharacterID: m.characterID,
			Type:        byte(invType),
			Slot:        slot,
			ItemID:      itemID,
			Quantity:    1,
		}
		
		if err := m.repo.Create(ctx(), item); err != nil {
			return nil, fmt.Errorf("failed to save item: %w", err)
		}
		
		m.cacheItem(item)
		operations = append(operations, &InventoryOperation{
			Type:     OpAdd,
			InvType:  invType,
			Slot:     slot,
			Item:     item,
			Quantity: 1,
		})
		
		log.Printf("[Inventory] Added equip %d to slot %d", itemID, slot)
		return operations, nil
	}
	
	// For stackable items, try to add to existing stacks first
	for remaining > 0 {
		// Find existing stack with space
		existingItem := m.findExistingStack(itemID, invType)
		if existingItem != nil {
			spaceInStack := maxStack - existingItem.Quantity
			toAdd := remaining
			if toAdd > spaceInStack {
				toAdd = spaceInStack
			}
			
			existingItem.Quantity += toAdd
			remaining -= toAdd
			
			if err := m.repo.Update(ctx(), existingItem); err != nil {
				return nil, fmt.Errorf("failed to update stack: %w", err)
			}
			
			operations = append(operations, &InventoryOperation{
				Type:     OpUpdateQuantity,
				InvType:  invType,
				Slot:     existingItem.Slot,
				Item:     existingItem,
				Quantity: existingItem.Quantity,
			})
			
			log.Printf("[Inventory] Updated stack at slot %d, now %d", existingItem.Slot, existingItem.Quantity)
			continue
		}
		
		// No existing stack, find new slot
		slot := m.findFreeSlot(invType)
		if slot == 0 {
			if len(operations) > 0 {
				// Partial success
				return operations, fmt.Errorf("inventory full, added %d of %d", quantity-remaining, quantity)
			}
			return nil, fmt.Errorf("inventory full")
		}
		
		toAdd := remaining
		if toAdd > maxStack {
			toAdd = maxStack
		}
		
		item := &models.Inventory{
			CharacterID: m.characterID,
			Type:        byte(invType),
			Slot:        slot,
			ItemID:      itemID,
			Quantity:    toAdd,
		}
		
		if err := m.repo.Create(ctx(), item); err != nil {
			return nil, fmt.Errorf("failed to save item: %w", err)
		}
		
		m.cacheItem(item)
		remaining -= toAdd
		
		operations = append(operations, &InventoryOperation{
			Type:     OpAdd,
			InvType:  invType,
			Slot:     slot,
			Item:     item,
			Quantity: toAdd,
		})
		
		log.Printf("[Inventory] Added new stack of %d x %d at slot %d", toAdd, itemID, slot)
	}
	
	return operations, nil
}

// RemoveItem removes items from inventory
func (m *Manager) RemoveItem(itemID int32, quantity int16) ([]*InventoryOperation, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("invalid quantity: %d", quantity)
	}
	
	invType := wz.GetInventoryTypeFromItemID(itemID)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil, fmt.Errorf("invalid inventory type")
	}
	
	var operations []*InventoryOperation
	remaining := quantity
	
	// Find items with this ID and remove
	for slot, item := range invMap {
		if item.ItemID != itemID {
			continue
		}
		
		if item.Quantity <= remaining {
			// Remove entire stack
			remaining -= item.Quantity
			
			if err := m.repo.Delete(ctx(), item); err != nil {
				return nil, fmt.Errorf("failed to delete item: %w", err)
			}
			
			m.removeFromCache(item)
			
			operations = append(operations, &InventoryOperation{
				Type:    OpRemove,
				InvType: invType,
				Slot:    slot,
			})
			
			log.Printf("[Inventory] Removed stack at slot %d", slot)
		} else {
			// Partial removal
			item.Quantity -= remaining
			remaining = 0
			
			if err := m.repo.Update(ctx(), item); err != nil {
				return nil, fmt.Errorf("failed to update item: %w", err)
			}
			
			operations = append(operations, &InventoryOperation{
				Type:     OpUpdateQuantity,
				InvType:  invType,
				Slot:     slot,
				Item:     item,
				Quantity: item.Quantity,
			})
			
			log.Printf("[Inventory] Updated stack at slot %d, now %d", slot, item.Quantity)
		}
		
		if remaining == 0 {
			break
		}
	}
	
	if remaining > 0 {
		return operations, fmt.Errorf("insufficient items, removed %d of %d", quantity-remaining, quantity)
	}
	
	return operations, nil
}

// HasItem checks if the character has at least the specified quantity of an item
func (m *Manager) HasItem(itemID int32, quantity int16) bool {
	if quantity <= 0 {
		return true
	}
	
	invType := wz.GetInventoryTypeFromItemID(itemID)
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return false
	}
	
	var total int16
	for _, item := range invMap {
		if item.ItemID == itemID {
			total += item.Quantity
		}
	}
	
	return total >= quantity
}

// GetItemCount returns the total quantity of an item
func (m *Manager) GetItemCount(itemID int32) int16 {
	invType := wz.GetInventoryTypeFromItemID(itemID)
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return 0
	}
	
	var total int16
	for _, item := range invMap {
		if item.ItemID == itemID {
			total += item.Quantity
		}
	}
	
	return total
}

// GetAllItems returns all items in all inventories
func (m *Manager) GetAllItems() []*models.Inventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var items []*models.Inventory
	
	for _, item := range m.equipped {
		items = append(items, item)
	}
	for _, item := range m.equip {
		items = append(items, item)
	}
	for _, item := range m.consume {
		items = append(items, item)
	}
	for _, item := range m.install {
		items = append(items, item)
	}
	for _, item := range m.etc {
		items = append(items, item)
	}
	for _, item := range m.cash {
		items = append(items, item)
	}
	
	return items
}

// GetItemsByType returns all items of a specific inventory type
func (m *Manager) GetItemsByType(invType models.InventoryType) []*models.Inventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil
	}
	
	items := make([]*models.Inventory, 0, len(invMap))
	for _, item := range invMap {
		items = append(items, item)
	}
	
	return items
}

// GetEquippedItems returns all equipped items (worn equipment)
func (m *Manager) GetEquippedItems() []*models.Inventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	items := make([]*models.Inventory, 0, len(m.equipped))
	for _, item := range m.equipped {
		items = append(items, item)
	}
	return items
}

// GetItemAtSlot returns the item at a specific slot
func (m *Manager) GetItemAtSlot(invType models.InventoryType, slot int16) *models.Inventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil
	}
	
	return invMap[slot]
}

// UseItem uses a consumable item at the specified slot
// Returns the operation performed and any stat changes to apply
func (m *Manager) UseItem(invType models.InventoryType, slot int16) (*InventoryOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil, fmt.Errorf("invalid inventory type")
	}
	
	item, exists := invMap[slot]
	if !exists {
		return nil, fmt.Errorf("no item at slot %d", slot)
	}
	
	// Decrease quantity
	item.Quantity--
	
	if item.Quantity <= 0 {
		// Remove item completely
		if err := m.repo.Delete(ctx(), item); err != nil {
			return nil, fmt.Errorf("failed to delete item: %w", err)
		}
		m.removeFromCache(item)
		
		log.Printf("[Inventory] Used and removed item %d at slot %d", item.ItemID, slot)
		return &InventoryOperation{
			Type:    OpRemove,
			InvType: invType,
			Slot:    slot,
			Item:    item,
		}, nil
	}
	
	// Update quantity
	if err := m.repo.Update(ctx(), item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	
	log.Printf("[Inventory] Used item %d at slot %d, remaining: %d", item.ItemID, slot, item.Quantity)
	return &InventoryOperation{
		Type:     OpUpdateQuantity,
		InvType:  invType,
		Slot:     slot,
		Item:     item,
		Quantity: item.Quantity,
	}, nil
}

// MoveItem moves an item from one slot to another within the same inventory
// If destSlot is 0, this is a drop operation
func (m *Manager) MoveItem(invType models.InventoryType, srcSlot, destSlot int16, quantity int16) ([]*InventoryOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil, fmt.Errorf("invalid inventory type")
	}
	
	srcItem, exists := invMap[srcSlot]
	if !exists {
		return nil, fmt.Errorf("no item at source slot %d", srcSlot)
	}
	
	var operations []*InventoryOperation
	
	// Drop operation (destSlot == 0)
	if destSlot == 0 {
		toRemove := quantity
		if toRemove <= 0 || toRemove > srcItem.Quantity {
			toRemove = srcItem.Quantity
		}
		
		srcItem.Quantity -= toRemove
		
		if srcItem.Quantity <= 0 {
			// Remove entire item
			if err := m.repo.Delete(ctx(), srcItem); err != nil {
				return nil, fmt.Errorf("failed to delete item: %w", err)
			}
			m.removeFromCache(srcItem)
			
			operations = append(operations, &InventoryOperation{
				Type:    OpRemove,
				InvType: invType,
				Slot:    srcSlot,
			})
			
			log.Printf("[Inventory] Dropped %d x item %d from slot %d", toRemove, srcItem.ItemID, srcSlot)
		} else {
			// Update quantity
			if err := m.repo.Update(ctx(), srcItem); err != nil {
				return nil, fmt.Errorf("failed to update item: %w", err)
			}
			
			operations = append(operations, &InventoryOperation{
				Type:     OpUpdateQuantity,
				InvType:  invType,
				Slot:     srcSlot,
				Item:     srcItem,
				Quantity: srcItem.Quantity,
			})
			
			log.Printf("[Inventory] Dropped %d x item %d from slot %d, remaining: %d", toRemove, srcItem.ItemID, srcSlot, srcItem.Quantity)
		}
		
		return operations, nil
	}
	
	// Check destination slot
	destItem, destExists := invMap[destSlot]
	
	if !destExists {
		// Simple move to empty slot
		delete(invMap, srcSlot)
		srcItem.Slot = destSlot
		invMap[destSlot] = srcItem
		
		if err := m.repo.Update(ctx(), srcItem); err != nil {
			return nil, fmt.Errorf("failed to update item: %w", err)
		}
		
		operations = append(operations, &InventoryOperation{
			Type:    OpMove,
			InvType: invType,
			Slot:    srcSlot,
			NewSlot: destSlot,
			Item:    srcItem,
		})
		
		log.Printf("[Inventory] Moved item %d from slot %d to slot %d", srcItem.ItemID, srcSlot, destSlot)
		return operations, nil
	}
	
	// Destination has an item - check if we can stack
	if srcItem.ItemID == destItem.ItemID && !wz.IsEquipItem(srcItem.ItemID) {
		dm := wz.GetInstance()
		maxStack := dm.GetSlotMax(srcItem.ItemID)
		
		if destItem.Quantity < maxStack {
			// Stack items
			spaceInDest := maxStack - destItem.Quantity
			toMove := srcItem.Quantity
			if toMove > spaceInDest {
				toMove = spaceInDest
			}
			
			destItem.Quantity += toMove
			srcItem.Quantity -= toMove
			
			if err := m.repo.Update(ctx(), destItem); err != nil {
				return nil, fmt.Errorf("failed to update dest item: %w", err)
			}
			
			operations = append(operations, &InventoryOperation{
				Type:     OpUpdateQuantity,
				InvType:  invType,
				Slot:     destSlot,
				Item:     destItem,
				Quantity: destItem.Quantity,
			})
			
			if srcItem.Quantity <= 0 {
				if err := m.repo.Delete(ctx(), srcItem); err != nil {
					return nil, fmt.Errorf("failed to delete src item: %w", err)
				}
				m.removeFromCache(srcItem)
				
				operations = append(operations, &InventoryOperation{
					Type:    OpRemove,
					InvType: invType,
					Slot:    srcSlot,
				})
			} else {
				if err := m.repo.Update(ctx(), srcItem); err != nil {
					return nil, fmt.Errorf("failed to update src item: %w", err)
				}
				
				operations = append(operations, &InventoryOperation{
					Type:     OpUpdateQuantity,
					InvType:  invType,
					Slot:     srcSlot,
					Item:     srcItem,
					Quantity: srcItem.Quantity,
				})
			}
			
			log.Printf("[Inventory] Stacked %d items from slot %d to slot %d", toMove, srcSlot, destSlot)
			return operations, nil
		}
	}
	
	// Swap items
	delete(invMap, srcSlot)
	delete(invMap, destSlot)
	
	srcItem.Slot = destSlot
	destItem.Slot = srcSlot
	
	invMap[destSlot] = srcItem
	invMap[srcSlot] = destItem
	
	if err := m.repo.Update(ctx(), srcItem); err != nil {
		return nil, fmt.Errorf("failed to update src item: %w", err)
	}
	if err := m.repo.Update(ctx(), destItem); err != nil {
		return nil, fmt.Errorf("failed to update dest item: %w", err)
	}
	
	operations = append(operations, &InventoryOperation{
		Type:    OpMove,
		InvType: invType,
		Slot:    srcSlot,
		NewSlot: destSlot,
		Item:    srcItem,
	})
	
	log.Printf("[Inventory] Swapped items: slot %d <-> slot %d", srcSlot, destSlot)
	return operations, nil
}

// GetItemBySlot returns the item at a specific slot (exported version)
func (m *Manager) GetItemBySlot(invType models.InventoryType, slot int16) *models.Inventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	invMap := m.getInventoryMap(invType)
	if invMap == nil {
		return nil
	}
	
	return invMap[slot]
}

// InventoryOperation represents a single inventory change
type InventoryOperation struct {
	Type     OperationType
	InvType  models.InventoryType
	Slot     int16
	Item     *models.Inventory // For add/update
	Quantity int16             // For update quantity
	NewSlot  int16             // For move operations
}

// OperationType represents the type of inventory operation
type OperationType byte

const (
	OpAdd            OperationType = 0
	OpUpdateQuantity OperationType = 1
	OpMove           OperationType = 2
	OpRemove         OperationType = 3
)

