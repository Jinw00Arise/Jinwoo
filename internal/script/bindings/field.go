package bindings

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/internal/game"
	"github.com/Jinw00Arise/Jinwoo/internal/packet"
	lua "github.com/yuin/gopher-lua"
)

// WarpFunc is a callback function to handle field transfers.
type WarpFunc func(session game.Session, mapID int32, portalName string)

// RegisterFieldBindings registers field-related Lua functions.
func RegisterFieldBindings(L *lua.LState, session game.Session, warpFunc WarpFunc) {
	// Warp to another map
	L.SetGlobal("warp", L.NewFunction(func(L *lua.LState) int {
		mapID := int32(L.CheckNumber(1))
		portalName := L.OptString(2, "sp")
		
		log.Printf("[Script] Warping %s to map %d (portal: %s)", 
			session.Character().GetName(), mapID, portalName)
		
		if warpFunc != nil {
			warpFunc(session, mapID, portalName)
		}
		return 0
	}))

	// Show balloon message
	L.SetGlobal("balloonMessage", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		width := int16(L.OptNumber(2, 150))
		duration := int16(L.OptNumber(3, 5))
		
		p := buildBalloonMessagePacket(text, width, duration)
		session.Send(p)
		
		return 0
	}))

	// Show avatar oriented effect
	L.SetGlobal("avatarOriented", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		
		p := buildAvatarOrientedPacket(path)
		session.Send(p)
		
		return 0
	}))

	// Send local effect
	L.SetGlobal("showEffect", L.NewFunction(func(L *lua.LState) int {
		effect := L.CheckString(1)
		
		p := buildLocalEffectPacket(effect)
		session.Send(p)
		
		return 0
	}))
}

// RegisterPortalBindings registers portal-specific Lua functions.
func RegisterPortalBindings(L *lua.LState, session game.Session, warpFunc WarpFunc) {
	RegisterFieldBindings(L, session, warpFunc)
	
	// Additional portal-specific functions can go here
}

func buildBalloonMessagePacket(text string, width, duration int16) packet.Packet {
	p := packet.NewWithOpcode(0x00F5) // SendUserBalloonMsg
	p.WriteString(text)
	p.WriteShort(uint16(width))
	p.WriteShort(uint16(duration))
	p.WriteBool(true) // avatar oriented
	return p
}

func buildAvatarOrientedPacket(path string) packet.Packet {
	p := packet.NewWithOpcode(0x012B) // SendUserEffectLocal
	p.WriteByte(0x0E)                  // Effect type: AvatarOriented
	p.WriteString(path)
	p.WriteInt(0) // duration (0 = default)
	return p
}

func buildLocalEffectPacket(effect string) packet.Packet {
	p := packet.NewWithOpcode(0x012B) // SendUserEffectLocal
	p.WriteByte(0x03)                  // Effect type: Quest
	p.WriteString(effect)
	return p
}

