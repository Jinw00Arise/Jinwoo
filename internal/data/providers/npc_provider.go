package providers

import (
	"log"
	"strconv"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

// NPCInfo contains NPC data loaded from String.wz
type NPCInfo struct {
	ID   int32
	Name string
	Func string // NPC function description (e.g., "Guide", "Shop")
}

// NPCProvider loads and caches NPC data from String.wz
type NPCProvider struct {
	wz    *wz.WzProvider
	npcs  map[int32]*NPCInfo
	mu    sync.RWMutex
	loaded bool
}

// NewNPCProvider creates a new NPC provider
func NewNPCProvider(wzProvider *wz.WzProvider) *NPCProvider {
	p := &NPCProvider{
		wz:   wzProvider,
		npcs: make(map[int32]*NPCInfo),
	}
	p.loadNPCStrings()
	return p
}

// GetNPCName returns the name of an NPC by template ID
func (p *NPCProvider) GetNPCName(npcID int32) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if info, ok := p.npcs[npcID]; ok {
		return info.Name
	}
	return ""
}

// GetNPCFunc returns the function description of an NPC
func (p *NPCProvider) GetNPCFunc(npcID int32) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if info, ok := p.npcs[npcID]; ok {
		return info.Func
	}
	return ""
}

// GetNPCInfo returns full NPC info
func (p *NPCProvider) GetNPCInfo(npcID int32) *NPCInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.npcs[npcID]
}

// loadNPCStrings loads NPC names from String.wz/Npc.img
func (p *NPCProvider) loadNPCStrings() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.loaded {
		return
	}

	stringDir := p.wz.Dir("String.wz")
	img, err := stringDir.Image("Npc")
	if err != nil {
		log.Printf("[NPCProvider] Failed to load String.wz/Npc.img: %v", err)
		return
	}

	root := img.Root()
	if root == nil {
		log.Printf("[NPCProvider] Npc.img has no root")
		return
	}

	// Each entry is named by NPC ID
	for i := range root.ImgDirs {
		npcDir := &root.ImgDirs[i]

		npcID, err := strconv.ParseInt(npcDir.Name, 10, 32)
		if err != nil {
			continue
		}

		info := &NPCInfo{
			ID: int32(npcID),
		}

		if name, err := npcDir.GetString("name"); err == nil {
			info.Name = name
		}

		if funcStr, err := npcDir.GetString("func"); err == nil {
			info.Func = funcStr
		}

		p.npcs[int32(npcID)] = info
	}

	p.loaded = true
	log.Printf("[NPCProvider] Loaded %d NPC strings", len(p.npcs))
}
