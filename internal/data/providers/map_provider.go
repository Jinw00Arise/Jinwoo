package providers

import (
	"fmt"
	"strconv"
)

type MapData struct {
	ID           int32
	ReturnMap    int32
	ForcedReturn int32
	SpawnPoint   Portal
	Portals      map[string]Portal
	Footholds    []Foothold
}

type Portal struct {
	Name string
	Type int32
	X    int16
	Y    int16
	TM   int32  // Target map
	TN   string // Target portal name
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
	wz *WzProvider
}

func NewMapProvider(wz *WzProvider) *MapProvider {
	return &MapProvider{wz: wz}
}

func (p *MapProvider) GetMapData(mapID int32) (*MapData, error) {
	// Determine map file path
	// Maps are organized like: Map/Map1/100000000.img.xml
	folder := fmt.Sprintf("Map%d", mapID/100000000)
	filename := fmt.Sprintf("%09d.img.xml", mapID)
	path := fmt.Sprintf("Map.wz/Map/%s/%s", folder, filename)

	// Load the IMG
	root, err := p.wz.GetImg(path)
	if err != nil {
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

	return p.parseMapData(mapID, root)
}

func (p *MapProvider) parseMapData(mapID int32, root *ImgDir) (*MapData, error) {
	mapData := &MapData{
		ID:      mapID,
		Portals: make(map[string]Portal),
	}

	// Parse info section
	if info := root.Get("info"); info != nil {
		mapData.ReturnMap = info.GetIntOrDefault("returnMap", mapID)
		mapData.ForcedReturn = info.GetIntOrDefault("forcedReturn", 999999999)
	}

	// Parse portals
	if portalSection := root.Get("portal"); portalSection != nil {
		for i := range portalSection.ImgDirs {
			portalDir := &portalSection.ImgDirs[i]
			portal := Portal{
				Name: portalDir.GetStringOrDefault("pn", ""),
				Type: portalDir.GetIntOrDefault("pt", 0),
				X:    int16(portalDir.GetIntOrDefault("x", 0)),
				Y:    int16(portalDir.GetIntOrDefault("y", 0)),
				TM:   portalDir.GetIntOrDefault("tm", 999999999),
				TN:   portalDir.GetStringOrDefault("tn", ""),
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
		mapData.Footholds = p.parseFootholds(footholdSection)
	}

	// If no spawn point found, use default
	if mapData.SpawnPoint.Name == "" {
		mapData.SpawnPoint = Portal{Name: "sp", X: 0, Y: 0}
	}

	return mapData, nil
}

func (p *MapProvider) parseFootholds(root *ImgDir) []Foothold {
	var footholds []Foothold

	// Footholds are nested: foothold/layer/group/foothold
	for _, layerDir := range root.ImgDirs {
		layer, _ := strconv.ParseInt(layerDir.Name, 10, 32)

		for _, groupDir := range layerDir.ImgDirs {
			for _, fhDir := range groupDir.ImgDirs {
				fhID, _ := strconv.ParseInt(fhDir.Name, 10, 32)

				fh := Foothold{
					ID:    int32(fhID),
					Layer: int32(layer),
					X1:    int16(fhDir.GetIntOrDefault("x1", 0)),
					Y1:    int16(fhDir.GetIntOrDefault("y1", 0)),
					X2:    int16(fhDir.GetIntOrDefault("x2", 0)),
					Y2:    int16(fhDir.GetIntOrDefault("y2", 0)),
					Prev:  fhDir.GetIntOrDefault("prev", 0),
					Next:  fhDir.GetIntOrDefault("next", 0),
				}
				footholds = append(footholds, fh)
			}
		}
	}

	return footholds
}
