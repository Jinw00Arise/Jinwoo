package channel

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/game/inventory"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

// handleUserChangeSlotPositionRequest handles item moving/dropping
func (h *Handler) handleUserChangeSlotPositionRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}

	_ = reader.ReadInt() // tTick (timestamp)
	invType := models.InventoryType(reader.ReadByte())
	srcSlot := int16(reader.ReadShort())
	destSlot := int16(reader.ReadShort())
	quantity := int16(reader.ReadShort())

	log.Printf("[Inventory] %s move/drop: type=%d src=%d dest=%d qty=%d",
		h.character().Name, invType, srcSlot, destSlot, quantity)

	// Check if this is a drop (destSlot = 0)
	if destSlot == 0 {
		// Get item info before removing
		item := h.inventory().GetItemBySlot(invType, srcSlot)
		if item == nil {
			log.Printf("[Inventory] Drop failed: no item at slot %d", srcSlot)
			h.conn.Write(EnableActionsPacket())
			return
		}

		itemID := item.ItemID
		dropQty := quantity
		if dropQty <= 0 || dropQty > item.Quantity {
			dropQty = item.Quantity
		}

		// Remove from inventory
		operations, err := h.inventory().MoveItem(invType, srcSlot, destSlot, quantity)
		if err != nil {
			log.Printf("[Inventory] Drop failed: %v", err)
			h.conn.Write(EnableActionsPacket())
			return
		}

		// Send inventory update
		if len(operations) > 0 {
			if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
				log.Printf("[Inventory] Failed to send inventory update: %v", err)
			}
		}

		// Spawn the drop on the ground
		h.spawnDrop(itemID, dropQty)
		return
	}

	// Handle the move
	operations, err := h.inventory().MoveItem(invType, srcSlot, destSlot, quantity)
	if err != nil {
		log.Printf("[Inventory] Move failed: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Send inventory updates to client
	if len(operations) > 0 {
		if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
			log.Printf("[Inventory] Failed to send inventory update: %v", err)
		}
	}
}

// spawnDrop creates a drop on the ground near the player
func (h *Handler) spawnDrop(itemID int32, quantity int16) {
	if h.user == nil {
		return
	}

	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	// Get player position
	posX, posY := h.user.Position()

	// Find the ground Y below the player
	groundY := currentStage.FindGroundY(posX, posY, posY)

	// Create drop using stage's drop manager (0 for questID = no quest restriction)
	drop := currentStage.AddDrop(itemID, quantity, posX, groundY, h.character().ID, 0)
	drop.SetStartPosition(posX, posY-game.DropHeight)

	log.Printf("[Drop] Spawned drop %d: item %d x%d at (%d, %d)", drop.ObjectID, itemID, quantity, drop.X, drop.Y)

	// Send drop packet to client
	// Type 0 (JUST_SHOWING) = instant appear at position (best for player drops)
	// Type 1 (CREATE) = falls from source to destination (for mob drops)
	// Type 3 (FADING_OUT) = disappears immediately (quest items)
	if err := h.conn.Write(DropEnterFieldPacket(drop, DropEnterCreate, posX, posY, 0)); err != nil {
		log.Printf("[Drop] Failed to send drop packet: %v", err)
	}
}

// handleDropPickUpRequest handles picking up dropped items
func (h *Handler) handleDropPickUpRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}

	currentStage := h.currentStage()
	if currentStage == nil {
		return
	}

	_ = reader.ReadByte()  // fieldKey
	_ = reader.ReadInt()   // tTick
	_ = reader.ReadShort() // ptPickup.x
	_ = reader.ReadShort() // ptPickup.y
	objectID := reader.ReadInt()
	_ = reader.ReadInt() // dwCrc

	log.Printf("[Drop] %s picking up drop %d", h.character().Name, objectID)

	// Find and remove the drop from stage
	drop := currentStage.Drops().Remove(uint32(objectID))
	if drop == nil {
		log.Printf("[Drop] Drop %d not found", objectID)
		h.conn.Write(EnableActionsPacket())
		return
	}

	char := h.character()

	if drop.IsMeso {
		// Meso pickup
		char.Meso += drop.Meso

		// Send meso stat update
		stats := map[uint32]int64{StatMeso: int64(char.Meso)}
		if err := h.conn.Write(StatChangedPacket(false, stats)); err != nil {
			log.Printf("[Drop] Failed to send meso stat change: %v", err)
		}

		// Send meso pickup message
		if err := h.conn.Write(MessagePickUpMesoPacket(drop.Meso)); err != nil {
			log.Printf("[Drop] Failed to send meso pickup message: %v", err)
		}

		log.Printf("[Drop] %s picked up %d meso", char.Name, drop.Meso)
	} else {
		// Item pickup
		operations, err := h.inventory().AddItem(drop.ItemID, drop.Quantity)
		if err != nil {
			log.Printf("[Drop] Failed to add item to inventory: %v", err)
			// Send pickup failed message
			h.conn.Write(EnableActionsPacket())
			return
		}

		// Send inventory update
		if len(operations) > 0 {
			if err := h.conn.Write(inventory.InventoryOperationPacket(operations, true)); err != nil {
				log.Printf("[Drop] Failed to send inventory update: %v", err)
			}
		}

		// Send pickup message
		if err := h.conn.Write(MessagePickUpItemPacket(drop.ItemID, int32(drop.Quantity))); err != nil {
			log.Printf("[Drop] Failed to send pickup message: %v", err)
		}

		log.Printf("[Drop] %s picked up item %d x%d", char.Name, drop.ItemID, drop.Quantity)
	}

	// Remove drop from field (broadcast to all users on stage)
	currentStage.Broadcast(DropLeaveFieldPacket(objectID, DropLeavePickUp, char.ID, 0))

	// Enable actions so player can continue playing
	h.conn.Write(EnableActionsPacket())
}

// handleUserStatChangeItemUseRequest handles using consumable items (HP/MP potions, etc.)
func (h *Handler) handleUserStatChangeItemUseRequest(reader *packet.Reader) {
	if h.user == nil || h.inventory() == nil {
		return
	}

	_ = reader.ReadInt() // tTick (timestamp)
	slot := int16(reader.ReadShort())
	itemID := reader.ReadInt()

	log.Printf("[Inventory] %s use item: slot=%d itemID=%d", h.character().Name, slot, itemID)

	// Verify the item exists at that slot
	item := h.inventory().GetItemBySlot(models.InventoryConsume, slot)
	if item == nil || item.ItemID != int32(itemID) {
		log.Printf("[Inventory] Item mismatch or not found at slot %d", slot)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Get item effects (HP/MP recovery amounts)
	hpRecover, mpRecover := h.getItemRecoveryAmounts(int32(itemID))

	// Use the item (decrements quantity)
	operation, err := h.inventory().UseItem(models.InventoryConsume, slot)
	if err != nil {
		log.Printf("[Inventory] Use item failed: %v", err)
		h.conn.Write(EnableActionsPacket())
		return
	}

	// Apply item effects
	statChanges := make(map[uint32]int64)

	if hpRecover > 0 {
		h.character().HP += hpRecover
		if h.character().HP > h.character().MaxHP {
			h.character().HP = h.character().MaxHP
		}
		statChanges[StatHP] = int64(h.character().HP)
		log.Printf("[Inventory] %s recovered %d HP (now %d/%d)",
			h.character().Name, hpRecover, h.character().HP, h.character().MaxHP)
	}

	if mpRecover > 0 {
		h.character().MP += mpRecover
		if h.character().MP > h.character().MaxMP {
			h.character().MP = h.character().MaxMP
		}
		statChanges[StatMP] = int64(h.character().MP)
		log.Printf("[Inventory] %s recovered %d MP (now %d/%d)",
			h.character().Name, mpRecover, h.character().MP, h.character().MaxMP)
	}

	// Send inventory update
	if operation != nil {
		if err := h.conn.Write(inventory.InventoryOperationPacket([]*inventory.InventoryOperation{operation}, true)); err != nil {
			log.Printf("[Inventory] Failed to send inventory update: %v", err)
		}
	}

	// Send stat update if HP/MP changed
	if len(statChanges) > 0 {
		if err := h.conn.Write(StatChangedPacket(true, statChanges)); err != nil {
			log.Printf("[Inventory] Failed to send stat update: %v", err)
		}
	}
}

// getItemRecoveryAmounts returns HP and MP recovery amounts for a consumable item
// Loads recovery data from WZ files (Item.wz/Consume/xxxx.img.xml -> spec node)
func (h *Handler) getItemRecoveryAmounts(itemID int32) (hpRecover, mpRecover int32) {
	dm := wz.GetInstance()
	if dm == nil {
		log.Printf("[Item] WZ DataManager not initialized!")
		return 0, 0
	}

	// Get item data from WZ
	hp, mp, hpR, mpR := dm.GetItemRecovery(itemID)

	// Debug: log the raw values
	log.Printf("[Item] WZ data for %d: hp=%d, mp=%d, hpR=%d, mpR=%d", itemID, hp, mp, hpR, mpR)

	// Calculate flat recovery
	hpRecover = hp
	mpRecover = mp

	// Add percentage-based recovery if applicable
	if hpR > 0 && h.character() != nil {
		hpRecover += (h.character().MaxHP * hpR) / 100
	}
	if mpR > 0 && h.character() != nil {
		mpRecover += (h.character().MaxMP * mpR) / 100
	}

	// Log the final recovery values
	log.Printf("[Item] %d final recovery: HP=%d, MP=%d", itemID, hpRecover, mpRecover)

	return hpRecover, mpRecover
}

// getInventoryData builds inventory data for SetField packet
func (h *Handler) getInventoryData() *InventoryData {
	inv := h.inventory()
	if inv == nil {
		return nil
	}

	return &InventoryData{
		Equipped: inv.GetItemsByType(models.InventoryEquipped),
		Equip:    inv.GetItemsByType(models.InventoryEquip),
		Consume:  inv.GetItemsByType(models.InventoryConsume),
		Install:  inv.GetItemsByType(models.InventoryInstall),
		Etc:      inv.GetItemsByType(models.InventoryEtc),
		Cash:     inv.GetItemsByType(models.InventoryCash),
	}
}

