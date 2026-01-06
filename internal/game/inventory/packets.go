package inventory

import (
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	"github.com/Jinw00Arise/Jinwoo/pkg/maple"
)

// InventoryOperationPacket creates a packet to update client inventory
func InventoryOperationPacket(operations []*InventoryOperation, exclRequest bool) packet.Packet {
	p := packet.NewWithOpcode(maple.SendInventoryOperation)
	p.WriteBool(exclRequest) // bExclRequest - re-enable actions if true
	p.WriteByte(byte(len(operations)))
	
	for _, op := range operations {
		p.WriteByte(byte(op.Type)) // Operation type (0=add, 1=update, 2=move, 3=remove)
		p.WriteByte(byte(op.InvType)) // Inventory type
		
		switch op.Type {
		case OpAdd:
			p.WriteShort(uint16(op.Slot))
			writeItem(&p, op.Item)
		case OpUpdateQuantity:
			p.WriteShort(uint16(op.Slot))
			p.WriteShort(uint16(op.Quantity))
		case OpMove:
			p.WriteShort(uint16(op.Slot))    // Old slot
			p.WriteShort(uint16(op.NewSlot)) // New slot
		case OpRemove:
			p.WriteShort(uint16(op.Slot))
		}
	}
	
	return p
}

// writeItem writes an item to the packet based on its type
func writeItem(p *packet.Packet, item *models.Inventory) {
	if item.IsEquip() {
		writeEquipItem(p, item)
	} else {
		writeStackableItem(p, item)
	}
}

// writeEquipItem writes an equipment item to the packet
func writeEquipItem(p *packet.Packet, item *models.Inventory) {
	p.WriteByte(1) // Item type (1 = equip)
	p.WriteInt(uint32(item.ItemID))
	p.WriteBool(false) // bCashItemSN (cash item unique ID flag)
	writeFileTime(p, time.Time{}) // ftExpire
	
	// Equipment stats
	slots := byte(7)
	if item.Slots != nil {
		slots = *item.Slots
	}
	p.WriteByte(slots) // nRUC (remaining upgrade count)
	
	level := byte(0)
	if item.Level != nil {
		level = *item.Level
	}
	p.WriteByte(level) // nCUC (current upgrade count)
	
	writeEquipStat(p, item.STR)
	writeEquipStat(p, item.DEX)
	writeEquipStat(p, item.INT)
	writeEquipStat(p, item.LUK)
	writeEquipStat(p, item.HP)
	writeEquipStat(p, item.MP)
	writeEquipStat(p, item.WAtk)
	writeEquipStat(p, item.MAtk)
	writeEquipStat(p, item.WDef)
	writeEquipStat(p, item.MDef)
	writeEquipStat(p, item.Accuracy)
	writeEquipStat(p, item.Avoidability)
	writeEquipStat(p, item.Hands)
	writeEquipStat(p, item.Speed)
	writeEquipStat(p, item.Jump)
	
	p.WriteString("")   // sTitle (owner name)
	p.WriteShort(0)     // nAttribute (item flags)
	p.WriteByte(0)      // nLevelUpType
	p.WriteByte(0)      // nLevel (item level)
	p.WriteInt(0)       // nEXP (item EXP)
	p.WriteInt(0xFFFFFFFF) // nDurability (-1)
	p.WriteInt(0)       // nIUC (vicious hammer)
	p.WriteByte(0)      // nGrade (potential grade)
	p.WriteByte(0)      // nCHUC (enhancement stars)
	p.WriteShort(0)     // nOption1 (potential line 1)
	p.WriteShort(0)     // nOption2 (potential line 2)
	p.WriteShort(0)     // nOption3 (potential line 3)
	p.WriteShort(0)     // nSocket1
	p.WriteShort(0)     // nSocket2
	p.WriteLong(0)      // liSN (serial number)
	writeFileTime(p, time.Time{}) // ftEquipped
	p.WriteInt(0xFFFFFFFF) // nPrevBonusExpRate (-1)
}

// writeStackableItem writes a stackable (non-equip) item to the packet
func writeStackableItem(p *packet.Packet, item *models.Inventory) {
	p.WriteByte(2) // Item type (2 = item/consume/etc)
	p.WriteInt(uint32(item.ItemID))
	p.WriteBool(false) // bCashItemSN
	writeFileTime(p, time.Time{}) // ftExpire
	p.WriteShort(uint16(item.Quantity))
	p.WriteString("")  // sTitle (owner)
	p.WriteShort(0)    // nAttribute (flags)
	
	// For rechargeable items (stars/bullets 207xxxx, 233xxxx)
	prefix := item.ItemID / 10000
	if prefix == 207 || prefix == 233 {
		p.WriteLong(0) // liSN (unique ID for rechargeable)
	}
}

// writeEquipStat writes a single stat value (or 0 if nil)
func writeEquipStat(p *packet.Packet, stat *int16) {
	if stat != nil {
		p.WriteShort(uint16(*stat))
	} else {
		p.WriteShort(0)
	}
}

// writeFileTime writes a Windows FILETIME (8 bytes)
func writeFileTime(p *packet.Packet, t time.Time) {
	if t.IsZero() {
		// Write default/permanent time
		p.WriteLong(150842304000000000)
		return
	}
	// Convert Unix time to Windows FILETIME
	const unixToFileTime = 116444736000000000
	ft := uint64(t.UnixNano()/100) + unixToFileTime
	p.WriteLong(ft)
}

// InventoryGrowPacket creates a packet to expand inventory size
func InventoryGrowPacket(invType models.InventoryType, newSize byte) packet.Packet {
	p := packet.NewWithOpcode(maple.SendInventoryGrow)
	p.WriteByte(byte(invType))
	p.WriteByte(newSize)
	return p
}

// Single item convenience functions

// AddItemPacket creates a packet to add a single item
func AddItemPacket(item *models.Inventory) packet.Packet {
	return InventoryOperationPacket([]*InventoryOperation{
		{
			Type:     OpAdd,
			InvType:  models.InventoryType(item.Type),
			Slot:     item.Slot,
			Item:     item,
			Quantity: item.Quantity,
		},
	}, true)
}

// UpdateItemQuantityPacket creates a packet to update item quantity
func UpdateItemQuantityPacket(invType models.InventoryType, slot int16, quantity int16) packet.Packet {
	return InventoryOperationPacket([]*InventoryOperation{
		{
			Type:     OpUpdateQuantity,
			InvType:  invType,
			Slot:     slot,
			Quantity: quantity,
		},
	}, true)
}

// RemoveItemPacket creates a packet to remove an item from a slot
func RemoveItemPacket(invType models.InventoryType, slot int16) packet.Packet {
	return InventoryOperationPacket([]*InventoryOperation{
		{
			Type:    OpRemove,
			InvType: invType,
			Slot:    slot,
		},
	}, true)
}

// MoveItemPacket creates a packet to move an item between slots
func MoveItemPacket(invType models.InventoryType, oldSlot, newSlot int16) packet.Packet {
	return InventoryOperationPacket([]*InventoryOperation{
		{
			Type:    OpMove,
			InvType: invType,
			Slot:    oldSlot,
			NewSlot: newSlot,
		},
	}, true)
}

