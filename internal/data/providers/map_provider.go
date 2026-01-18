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
	portal.X = int16(x)

	y, err := portalDir.GetInt("y")
	if err != nil {
		return portal, fmt.Errorf("missing y: %w", err)
	}
	portal.Y = int16(y)

	// tm and tn are optional (not all portals have targets)
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
