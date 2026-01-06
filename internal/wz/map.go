package wz

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// MapLife represents an NPC or mob spawn on a map
type MapLife struct {
	ID       int    // NPC or Mob ID
	Type     string // "n" for NPC, "m" for mob
	X        int    // X position
	Y        int    // Y position
	Cy       int    // Foothold Y
	FH       int    // Foothold ID
	RX0      int    // Left bound
	RX1      int    // Right bound
	MobTime  int    // Respawn time for mobs
	F        int    // Facing direction (0 = right, 1 = left)
	Hide     bool   // Hidden NPC
}

// MapPortal represents a portal on a map
type MapPortal struct {
	ID       int    // Portal ID (index)
	Name     string // Portal name
	Type     int    // Portal type (0=start, 1=visible, 2=hidden, etc.)
	X        int    // X position
	Y        int    // Y position
	ToMap    int    // Destination map ID
	ToName   string // Destination portal name
	Script   string // Portal script name
}

// MapFoothold represents a foothold (platform) on a map
type MapFoothold struct {
	ID    int
	X1    int
	Y1    int
	X2    int
	Y2    int
	Next  int
	Prev  int
	Layer int
}

// MapData contains all data for a single map
type MapData struct {
	ID        int
	Name      string
	ReturnMap int
	ForcedReturn int
	MobRate   float64
	NPCs      []MapLife
	Mobs      []MapLife
	Portals   []MapPortal
	Footholds []MapFoothold
}

// LoadMapData loads map data from a WZ XML file
func LoadMapData(wzPath string, mapID int) (*MapData, error) {
	// Determine the map category (Map0, Map1, Map2, etc.)
	category := mapID / 100000000
	fileName := fmt.Sprintf("%09d.img.xml", mapID)
	filePath := filepath.Join(wzPath, "Map.wz", "Map", fmt.Sprintf("Map%d", category), fileName)

	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	mapData := &MapData{
		ID:      mapID,
		MobRate: 1.0,
	}

	// Parse info section
	if info := root.GetChild("info"); info != nil {
		mapData.ReturnMap = info.GetInt("returnMap")
		mapData.ForcedReturn = info.GetInt("forcedReturn")
		if mobRate := info.GetFloat("mobRate"); mobRate > 0 {
			mapData.MobRate = mobRate
		}
	}

	// Parse life section (NPCs and mobs)
	if life := root.GetChild("life"); life != nil {
		for _, lifeNode := range life.GetAllChildren() {
			ml := MapLife{
				ID:      parseID(lifeNode.GetString("id")),
				Type:    lifeNode.GetString("type"),
				X:       lifeNode.GetInt("x"),
				Y:       lifeNode.GetInt("y"),
				Cy:      lifeNode.GetInt("cy"),
				FH:      lifeNode.GetInt("fh"),
				RX0:     lifeNode.GetInt("rx0"),
				RX1:     lifeNode.GetInt("rx1"),
				MobTime: lifeNode.GetInt("mobTime"),
				F:       lifeNode.GetInt("f"),
				Hide:    lifeNode.GetInt("hide") == 1,
			}

			if ml.Type == "n" {
				mapData.NPCs = append(mapData.NPCs, ml)
			} else if ml.Type == "m" {
				mapData.Mobs = append(mapData.Mobs, ml)
			}
		}
	}

	// Parse portal section
	if portal := root.GetChild("portal"); portal != nil {
		for i, portalNode := range portal.GetAllChildren() {
			p := MapPortal{
				ID:     i,
				Name:   portalNode.GetString("pn"),
				Type:   portalNode.GetInt("pt"),
				X:      portalNode.GetInt("x"),
				Y:      portalNode.GetInt("y"),
				ToMap:  portalNode.GetInt("tm"),
				ToName: portalNode.GetString("tn"),
				Script: portalNode.GetString("script"),
			}
			mapData.Portals = append(mapData.Portals, p)
		}
	}

	// Parse foothold section
	if foothold := root.GetChild("foothold"); foothold != nil {
		for _, layerNode := range foothold.GetAllChildren() {
			layer, _ := strconv.Atoi(layerNode.Name)
			for _, groupNode := range layerNode.GetAllChildren() {
				for _, fhNode := range groupNode.GetAllChildren() {
					fhID, _ := strconv.Atoi(fhNode.Name)
					fh := MapFoothold{
						ID:    fhID,
						X1:    fhNode.GetInt("x1"),
						Y1:    fhNode.GetInt("y1"),
						X2:    fhNode.GetInt("x2"),
						Y2:    fhNode.GetInt("y2"),
						Next:  fhNode.GetInt("next"),
						Prev:  fhNode.GetInt("prev"),
						Layer: layer,
					}
					mapData.Footholds = append(mapData.Footholds, fh)
				}
			}
		}
	}

	return mapData, nil
}

// parseID parses a string ID that may have leading zeros
func parseID(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}

