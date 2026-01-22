package server

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
	"github.com/Jinw00Arise/Jinwoo/internal/game/packets"
	"github.com/Jinw00Arise/Jinwoo/internal/game/script"
	"github.com/Jinw00Arise/Jinwoo/internal/protocol"
)

// ScriptCharacter wraps field.Character to implement script.CharacterAccessor
// It provides access to channel for field resolution
type ScriptCharacter struct {
	*field.Character
	channel *Channel
	client  *Client
}

// NewScriptCharacter creates a new script character adapter
func NewScriptCharacter(char *field.Character, channel *Channel, client *Client) *ScriptCharacter {
	return &ScriptCharacter{
		Character: char,
		channel:   channel,
		client:    client,
	}
}

// Ensure ScriptCharacter implements CharacterAccessor
var _ script.CharacterAccessor = (*ScriptCharacter)(nil)

// TransferField warps the character to a different map
func (sc *ScriptCharacter) TransferField(targetMapID int32, portalName string) {
	if sc.channel == nil {
		log.Printf("[Script] Cannot warp - no channel reference")
		return
	}

	targetField, err := sc.channel.GetField(targetMapID)
	if err != nil {
		log.Printf("[Script] Failed to get target field %d: %v", targetMapID, err)
		return
	}

	// Use the character's transfer method
	sc.Character.TransferToField(targetField, portalName)

	// Send SetField packet
	if sc.client != nil {
		items := sc.Character.Items()
		quests := sc.Character.QuestRecords()
		if err := sc.client.Write(SetField(sc.Character.Model(), int(sc.channel.ID()), sc.Character.FieldKey(), items, quests)); err != nil {
			log.Printf("[Script] Failed to send SetField after warp: %v", err)
		}

		// Send field entities (NPCs, mobs, other characters)
		sc.sendFieldEntities(targetField)

		// Enable actions after field transfer
		sc.client.Write(packets.EnableActions())
	}

	log.Printf("[Script] Warped %s to map %d (portal: %s)", sc.Character.Name(), targetMapID, portalName)
}

// sendFieldEntities sends all NPCs, mobs, and other characters in a field to the character.
func (sc *ScriptCharacter) sendFieldEntities(targetField *field.Field) {
	server := sc.client.server

	// Send NPCs
	npcProvider := server.NPCProvider()
	for _, npc := range targetField.GetAllNPCs() {
		sc.Character.Write(packets.NpcEnterField(npc))
		if npcProvider != nil {
			npcName := npcProvider.GetNPCName(npc.TemplateID())
			if npcName != "" {
				log.Printf("Spawned NPC: %s (id=%d, obj=%d) at (%d, %d)", npcName, npc.TemplateID(), npc.ObjectID(), npc.GetX(), npc.GetY())
			}
		}
	}
	targetField.AssignControllerToNPCs(sc.Character)

	// Send mobs
	for _, mob := range targetField.GetAliveMobs() {
		sc.Character.Write(packets.MobEnterField(mob))
		log.Printf("Spawned Mob: id=%d, obj=%d at (%d, %d)", mob.TemplateID(), mob.ObjectID(), mob.GetX(), mob.GetY())
	}
	targetField.AssignControllerToMobs(sc.Character)

	// Send other characters
	for _, otherChar := range targetField.GetAllCharacters() {
		if otherChar.ID() != sc.Character.ID() {
			sc.Character.Write(UserEnterField(otherChar))
		}
	}

	// Broadcast entry to others
	targetField.BroadcastExcept(UserEnterField(sc.Character), sc.Character)
}

// Write sends a packet to the character
func (sc *ScriptCharacter) Write(p protocol.Packet) error {
	return sc.Character.Write(p)
}
