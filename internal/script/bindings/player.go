// Package bindings provides Lua bindings for game functions.
package bindings

import (
	"github.com/Jinw00Arise/Jinwoo/internal/game"
	lua "github.com/yuin/gopher-lua"
)

// RegisterPlayerBindings registers player-related Lua functions.
func RegisterPlayerBindings(L *lua.LState, session game.Session) {
	// Character info getters
	L.SetGlobal("getCharID", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetID()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getName", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LString(session.Character().GetName()))
		} else {
			L.Push(lua.LString(""))
		}
		return 1
	}))

	L.SetGlobal("getLevel", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetLevel()))
		} else {
			L.Push(lua.LNumber(1))
		}
		return 1
	}))

	L.SetGlobal("getJob", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetJob()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getGender", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetGender()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getMapId", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetMapID()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	// Stats
	L.SetGlobal("getHp", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetHP()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getMaxHp", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetMaxHP()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getMp", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetMP()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getMaxMp", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetMaxMP()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getExp", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetEXP()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getMeso", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetMeso()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	L.SetGlobal("getFame", L.NewFunction(func(L *lua.LState) int {
		if session.Character() != nil {
			L.Push(lua.LNumber(session.Character().GetFame()))
		} else {
			L.Push(lua.LNumber(0))
		}
		return 1
	}))

	// Stat setters
	L.SetGlobal("setHp", L.NewFunction(func(L *lua.LState) int {
		hp := int32(L.CheckNumber(1))
		if session.Character() != nil {
			session.Character().SetHP(hp)
			// TODO: Send stat changed packet
		}
		return 0
	}))

	L.SetGlobal("setMp", L.NewFunction(func(L *lua.LState) int {
		mp := int32(L.CheckNumber(1))
		if session.Character() != nil {
			session.Character().SetMP(mp)
			// TODO: Send stat changed packet
		}
		return 0
	}))
}

