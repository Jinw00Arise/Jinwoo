package data

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MapData holds essential map information
type MapData struct {
	ID         int32
	ReturnMap  int32
	SpawnPoint Portal
	Portals    map[string]Portal
}

// Portal represents a spawn point or portal
type Portal struct {
	Name string
	X    int16
	Y    int16
	TM   int32 // Target map
}

// Manager handles loading and caching game data from WZ files
type Manager struct {
	wzPath string
	maps   map[int32]*MapData
	mu     sync.RWMutex
}

// NewManager creates a new data manager
func NewManager(wzPath string) *Manager {
	return &Manager{
		wzPath: wzPath,
		maps:   make(map[int32]*MapData),
	}
}

// GetMapData loads map data from WZ files (cached)
func (m *Manager) GetMapData(mapID int32) (*MapData, error) {
	// Fast path: check cache
	m.mu.RLock()
	if mapData, exists := m.maps[mapID]; exists {
		m.mu.RUnlock()
		return mapData, nil
	}
	m.mu.RUnlock()

	// Slow path: load from file
	mapData, err := m.loadMapData(mapID)
	if err != nil {
		return nil, err
	}

	// Cache it
	m.mu.Lock()
	m.maps[mapID] = mapData
	m.mu.Unlock()

	return mapData, nil
}

// loadMapData parses map XML file
func (m *Manager) loadMapData(mapID int32) (*MapData, error) {
	// Determine map file path
	// Maps are organized like: Map/Map1/100000000.img.xml
	folder := fmt.Sprintf("Map%d", mapID/100000000)
	filename := fmt.Sprintf("%09d.img.xml", mapID)
	path := filepath.Join(m.wzPath, "Map.wz", "Map", folder, filename)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return default map data for missing maps
		return &MapData{
			ID:        mapID,
			ReturnMap: mapID,
			SpawnPoint: Portal{
				Name: "sp",
				X:    0,
				Y:    0,
			},
			Portals: make(map[string]Portal),
		}, nil
	}

	// Parse XML
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file %s: %w", path, err)
	}
	defer file.Close()

	var root ImgDir
	if err := xml.NewDecoder(file).Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to parse map XML %s: %w", path, err)
	}

	return parseMapData(mapID, &root)
}

// parseMapData extracts relevant data from parsed XML
func parseMapData(mapID int32, root *ImgDir) (*MapData, error) {
	mapData := &MapData{
		ID:      mapID,
		Portals: make(map[string]Portal),
	}

	// Find info section
	for i := range root.ImgDirs {
		dir := &root.ImgDirs[i]
		switch dir.Name {
		case "info":
			mapData.ReturnMap = findIntValue(dir, "returnMap", mapID)

		case "portal":
			// Parse all portals
			for j := range dir.ImgDirs {
				portalDir := &dir.ImgDirs[j]
				portal := Portal{
					Name: findStringValue(portalDir, "pn", ""),
					X:    int16(findIntValue(portalDir, "x", 0)),
					Y:    int16(findIntValue(portalDir, "y", 0)),
					TM:   findIntValue(portalDir, "tm", 999999999),
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
	}

	// If no spawn point found, use default
	if mapData.SpawnPoint.Name == "" {
		mapData.SpawnPoint = Portal{Name: "sp", X: 0, Y: 0}
	}

	return mapData, nil
}

// XML structure for WZ files
type ImgDir struct {
	XMLName xml.Name  `xml:"imgdir"`
	Name    string    `xml:"name,attr"`
	ImgDirs []ImgDir  `xml:"imgdir"`
	Ints    []IntNode `xml:"int"`
	Strings []StrNode `xml:"string"`
}

type IntNode struct {
	Name  string `xml:"name,attr"`
	Value int32  `xml:"value,attr"`
}

type StrNode struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Helper functions to find values in XML tree
func findIntValue(dir *ImgDir, name string, defaultVal int32) int32 {
	for _, intNode := range dir.Ints {
		if intNode.Name == name {
			return intNode.Value
		}
	}
	return defaultVal
}

func findStringValue(dir *ImgDir, name string, defaultVal string) string {
	for _, strNode := range dir.Strings {
		if strNode.Name == name {
			return strNode.Value
		}
	}
	return defaultVal
}
