package providers

import (
	"fmt"
	"strconv"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

type MapData struct {
	ID           int32
	ReturnMap    int32
	ForcedReturn int32
	SpawnPoint   Portal
	Portals      map[string]Portal
	Footholds    []Foothold
	NPCSpawns    []LifeSpawn
	MobSpawns    []LifeSpawn
}

// LifeType indicates whether a life entry is an NPC or mob
type LifeType string

const (
	LifeTypeNPC LifeType = "n"
	LifeTypeMob LifeType = "m"
)

// LifeSpawn represents an NPC or mob spawn point from map WZ data
type LifeSpawn struct {
	Type    LifeType
	ID      int32 // Template ID (NPC ID or Mob ID)
	X       uint16
	Y       uint16
	Cy      uint16 // Spawn Y (often same as Y)
	Fh      uint16 // Foothold
	Rx0     uint16 // Left roaming bound
	Rx1     uint16 // Right roaming bound
	F       bool   // Flipped (facing left)
	Hide    bool   // Hidden on spawn
	MobTime int32  // Respawn time for mobs (milliseconds)
	Team    int32  // Team number (for PvP maps)
}

// Portal type constants
const (
	PortalTypeStartPoint int32 = 0 // Spawn point
	PortalTypeVisible    int32 = 1 // Regular portal
	PortalTypeHidden     int32 = 2 // Hidden portal (activated by proximity)
	PortalTypeScripted   int32 = 3 // Script-triggered portal
	PortalTypeAutomatic  int32 = 4 // Auto-trigger on contact
	PortalTypeCollision  int32 = 6 // Collision-based portal
	PortalTypeChangeable int32 = 7 // Can change destination
)

// PortalInvalidTarget is the sentinel value for portals without a target map
const PortalInvalidTarget int32 = 999999999

type Portal struct {
	Name   string
	Type   int32
	X      uint16
	Y      uint16
	TM     int32  // Target map (PortalInvalidTarget if none)
	TN     string // Target portal name
	Script string
}

// HasTarget returns true if this portal has a valid target map
func (p Portal) HasTarget() bool {
	return p.TM != PortalInvalidTarget && p.TM != 0
}

// IsUsableByPlayer returns true if this portal type can be activated by player input
func (p Portal) IsUsableByPlayer() bool {
	switch p.Type {
	case PortalTypeVisible, PortalTypeHidden, PortalTypeAutomatic:
		return true
	default:
		return false
	}
}

func (p Portal) GetScript() string {
	return p.Script
}

type Foothold struct {
	ID    int32
	Layer int32
	X1    int16
	Y1    int16
	X2    int16
	Y2    int16
	Prev  int32
	Next  int32
}

type MapProvider struct {
	wz *wz.WzProvider
}

func NewMapProvider(wzProvider *wz.WzProvider) *MapProvider {
	return &MapProvider{wz: wzProvider}
}

func (p *MapProvider) GetMapData(mapID int32) (*MapData, error) {
	// Maps are organized like: Map.wz/Map/Map1/100000000.img.xml
	folder := fmt.Sprintf("Map%d", mapID/100000000)
	filename := fmt.Sprintf("%09d", mapID)

	mapDir := p.wz.Dir("Map.wz").Dir("Map").Dir(folder)
	img, err := mapDir.Image(filename)
	if err != nil {
		return nil, fmt.Errorf("could not load map %d: %w", mapID, err)
	}

	root := img.Root()
	if root == nil {
		return nil, fmt.Errorf("map %d has no root", mapID)
	}

	return p.parseMapData(mapID, root)
}

func (p *MapProvider) parseMapData(mapID int32, root *wz.ImgDir) (*MapData, error) {
	mapData := &MapData{
		ID:      mapID,
		Portals: make(map[string]Portal),
	}

	// Parse info section
	info := root.Get("info")
	if info == nil {
		return nil, fmt.Errorf("map %d missing info section", mapID)
	}

	returnMap, err := info.GetInt("returnMap")
	if err != nil {
		return nil, fmt.Errorf("map %d: %w", mapID, err)
	}
	mapData.ReturnMap = returnMap

	forcedReturn, err := info.GetInt("forcedReturn")
	if err != nil {
		// forcedReturn might be optional, use returnMap as fallback
		mapData.ForcedReturn = returnMap
	} else {
		mapData.ForcedReturn = forcedReturn
	}

	// Parse portals
	if portalSection := root.Get("portal"); portalSection != nil {
		for i := range portalSection.ImgDirs {
			portalDir := &portalSection.ImgDirs[i]

			portal, err := p.parsePortal(portalDir)
			if err != nil {
				return nil, fmt.Errorf("map %d portal %s: %w", mapID, portalDir.Name, err)
			}

			// First "sp" portal is the spawn point
			if portal.Name == "sp" && mapData.SpawnPoint.Name == "" {
				mapData.SpawnPoint = portal
			}

			if portal.Name != "" {
				mapData.Portals[portal.Name] = portal
			}
		}
	}

	// Parse footholds
	if footholdSection := root.Get("foothold"); footholdSection != nil {
		footholds, err := p.parseFootholds(footholdSection)
		if err != nil {
			return nil, fmt.Errorf("map %d footholds: %w", mapID, err)
		}
		mapData.Footholds = footholds
	}

	// Parse life (NPCs and mobs)
	if lifeSection := root.Get("life"); lifeSection != nil {
		npcs, mobs, err := p.parseLife(lifeSection)
		if err != nil {
			return nil, fmt.Errorf("map %d life: %w", mapID, err)
		}
		mapData.NPCSpawns = npcs
		mapData.MobSpawns = mobs
	}

	return mapData, nil
}

func (p *MapProvider) parsePortal(portalDir *wz.ImgDir) (Portal, error) {
	portal := Portal{}

	pn, err := portalDir.GetString("pn")
	if err != nil {
		return portal, fmt.Errorf("missing pn: %w", err)
	}
	portal.Name = pn

	pt, err := portalDir.GetInt("pt")
	if err != nil {
		return portal, fmt.Errorf("missing pt: %w", err)
	}
	portal.Type = pt

	x, err := portalDir.GetInt("x")
	if err != nil {
		return portal, fmt.Errorf("missing x: %w", err)
	}
	portal.X = uint16(x)

	y, err := portalDir.GetInt("y")
	if err != nil {
		return portal, fmt.Errorf("missing y: %w", err)
	}
	portal.Y = uint16(y)

	script, err := portalDir.GetString("script")
	if script != "" {
		portal.Script = script
	}

	// tm and tn are optional (not all portals have targets)
	// Use sentinel value for missing tm to prevent accidental teleport to map 0
	portal.TM = PortalInvalidTarget
	if tm, err := portalDir.GetInt("tm"); err == nil {
		portal.TM = tm
	}
	if tn, err := portalDir.GetString("tn"); err == nil {
		portal.TN = tn
	}

	return portal, nil
}

func (p *MapProvider) parseFootholds(root *wz.ImgDir) ([]Foothold, error) {
	var footholds []Foothold

	// Footholds are nested: foothold/layer/group/foothold
	for i := range root.ImgDirs {
		layerDir := &root.ImgDirs[i]
		layer, err := strconv.ParseInt(layerDir.Name, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid layer name %s: %w", layerDir.Name, err)
		}

		for j := range layerDir.ImgDirs {
			groupDir := &layerDir.ImgDirs[j]

			for k := range groupDir.ImgDirs {
				fhDir := &groupDir.ImgDirs[k]
				fhID, err := strconv.ParseInt(fhDir.Name, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid foothold id %s: %w", fhDir.Name, err)
				}

				fh, err := p.parseFoothold(fhDir, int32(fhID), int32(layer))
				if err != nil {
					return nil, fmt.Errorf("foothold %d: %w", fhID, err)
				}
				footholds = append(footholds, fh)
			}
		}
	}

	return footholds, nil
}

func (p *MapProvider) parseFoothold(fhDir *wz.ImgDir, id, layer int32) (Foothold, error) {
	fh := Foothold{
		ID:    id,
		Layer: layer,
	}

	x1, err := fhDir.GetInt("x1")
	if err != nil {
		return fh, fmt.Errorf("missing x1: %w", err)
	}
	fh.X1 = int16(x1)

	y1, err := fhDir.GetInt("y1")
	if err != nil {
		return fh, fmt.Errorf("missing y1: %w", err)
	}
	fh.Y1 = int16(y1)

	x2, err := fhDir.GetInt("x2")
	if err != nil {
		return fh, fmt.Errorf("missing x2: %w", err)
	}
	fh.X2 = int16(x2)

	y2, err := fhDir.GetInt("y2")
	if err != nil {
		return fh, fmt.Errorf("missing y2: %w", err)
	}
	fh.Y2 = int16(y2)

	prev, err := fhDir.GetInt("prev")
	if err != nil {
		return fh, fmt.Errorf("missing prev: %w", err)
	}
	fh.Prev = prev

	next, err := fhDir.GetInt("next")
	if err != nil {
		return fh, fmt.Errorf("missing next: %w", err)
	}
	fh.Next = next

	return fh, nil
}

func (p *MapProvider) parseLife(lifeSection *wz.ImgDir) (npcs, mobs []LifeSpawn, err error) {
	for i := range lifeSection.ImgDirs {
		lifeDir := &lifeSection.ImgDirs[i]

		spawn, err := p.parseLifeSpawn(lifeDir)
		if err != nil {
			// Skip invalid life entries rather than failing the whole map
			continue
		}

		switch spawn.Type {
		case LifeTypeNPC:
			npcs = append(npcs, spawn)
		case LifeTypeMob:
			mobs = append(mobs, spawn)
		}
	}

	return npcs, mobs, nil
}

func (p *MapProvider) parseLifeSpawn(lifeDir *wz.ImgDir) (LifeSpawn, error) {
	spawn := LifeSpawn{}

	// Type is required: "n" for NPC, "m" for mob
	typeStr, err := lifeDir.GetString("type")
	if err != nil {
		return spawn, fmt.Errorf("missing type: %w", err)
	}
	spawn.Type = LifeType(typeStr)

	// ID is the template ID - can be string in WZ
	idStr, err := lifeDir.GetString("id")
	if err != nil {
		return spawn, fmt.Errorf("missing id: %w", err)
	}
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return spawn, fmt.Errorf("invalid id %s: %w", idStr, err)
	}
	spawn.ID = int32(id)

	// Position - required
	x, err := lifeDir.GetInt("x")
	if err != nil {
		return spawn, fmt.Errorf("missing x: %w", err)
	}
	spawn.X = uint16(x)

	y, err := lifeDir.GetInt("y")
	if err != nil {
		return spawn, fmt.Errorf("missing y: %w", err)
	}
	spawn.Y = uint16(y)

	// Optional fields with defaults
	if cy, err := lifeDir.GetInt("cy"); err == nil {
		spawn.Cy = uint16(cy)
	} else {
		spawn.Cy = spawn.Y // Default to Y
	}

	if fh, err := lifeDir.GetInt("fh"); err == nil {
		spawn.Fh = uint16(fh)
	}

	if rx0, err := lifeDir.GetInt("rx0"); err == nil {
		spawn.Rx0 = uint16(rx0)
	} else {
		spawn.Rx0 = spawn.X // Default to X
	}

	if rx1, err := lifeDir.GetInt("rx1"); err == nil {
		spawn.Rx1 = uint16(rx1)
	} else {
		spawn.Rx1 = spawn.X // Default to X
	}

	// f = 1 means flipped (facing left)
	if f, err := lifeDir.GetInt("f"); err == nil {
		spawn.F = f == 1
	}

	// hide = 1 means hidden
	if hide, err := lifeDir.GetInt("hide"); err == nil {
		spawn.Hide = hide == 1
	}

	// mobTime - respawn time for mobs
	if mobTime, err := lifeDir.GetInt("mobTime"); err == nil {
		spawn.MobTime = mobTime
	}

	// team - for PvP maps
	if team, err := lifeDir.GetInt("team"); err == nil {
		spawn.Team = team
	}

	return spawn, nil
}
