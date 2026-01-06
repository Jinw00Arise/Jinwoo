package wz

import (
	"path/filepath"
	"strconv"
)

// NPCStrings holds NPC name data loaded from String.wz
type NPCStrings struct {
	Names map[int]string // NPC ID -> Name
	Funcs map[int]string // NPC ID -> Function description
}

// LoadNPCStrings loads NPC names from String.wz/Npc.img.xml
func LoadNPCStrings(wzPath string) (*NPCStrings, error) {
	filePath := filepath.Join(wzPath, "String.wz", "Npc.img.xml")

	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	strings := &NPCStrings{
		Names: make(map[int]string),
		Funcs: make(map[int]string),
	}

	// Each child is an NPC entry with name as the NPC ID
	for _, npcNode := range root.GetAllChildren() {
		npcID, err := strconv.Atoi(npcNode.Name)
		if err != nil {
			continue
		}

		if name := npcNode.GetString("name"); name != "" {
			strings.Names[npcID] = name
		}
		if funcDesc := npcNode.GetString("func"); funcDesc != "" {
			strings.Funcs[npcID] = funcDesc
		}
	}

	return strings, nil
}

// MobStrings holds mob name data loaded from String.wz
type MobStrings struct {
	Names map[int]string // Mob ID -> Name
}

// LoadMobStrings loads mob names from String.wz/Mob.img.xml
func LoadMobStrings(wzPath string) (*MobStrings, error) {
	filePath := filepath.Join(wzPath, "String.wz", "Mob.img.xml")

	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	strings := &MobStrings{
		Names: make(map[int]string),
	}

	for _, mobNode := range root.GetAllChildren() {
		mobID, err := strconv.Atoi(mobNode.Name)
		if err != nil {
			continue
		}

		if name := mobNode.GetString("name"); name != "" {
			strings.Names[mobID] = name
		}
	}

	return strings, nil
}

// MapStrings holds map name data loaded from String.wz
type MapStrings struct {
	Names       map[int]string // Map ID -> Map name
	StreetNames map[int]string // Map ID -> Street name
}

// LoadMapStrings loads map names from String.wz/Map.img.xml
func LoadMapStrings(wzPath string) (*MapStrings, error) {
	filePath := filepath.Join(wzPath, "String.wz", "Map.img.xml")

	root, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	strings := &MapStrings{
		Names:       make(map[int]string),
		StreetNames: make(map[int]string),
	}

	// Maps are organized by region
	for _, regionNode := range root.GetAllChildren() {
		for _, mapNode := range regionNode.GetAllChildren() {
			mapID, err := strconv.Atoi(mapNode.Name)
			if err != nil {
				continue
			}

			if name := mapNode.GetString("mapName"); name != "" {
				strings.Names[mapID] = name
			}
			if street := mapNode.GetString("streetName"); street != "" {
				strings.StreetNames[mapID] = street
			}
		}
	}

	return strings, nil
}

