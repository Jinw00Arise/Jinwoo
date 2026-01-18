package item

import (
	"strconv"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

type ItemOptionLevelData struct {
	Level int32
	Props map[string]int32
}

func NewItemOptionLevelData(level int32, levelDir *wz.ImgDir) *ItemOptionLevelData {
	data := &ItemOptionLevelData{
		Level: level,
		Props: make(map[string]int32),
	}

	if levelDir == nil {
		return data
	}

	// Parse all int properties
	for _, intNode := range levelDir.Ints {
		data.Props[intNode.Name] = intNode.Value
	}

	return data
}

func (d *ItemOptionLevelData) GetProp(name string) (int32, bool) {
	v, ok := d.Props[name]
	return v, ok
}

func (d *ItemOptionLevelData) GetPropOr(name string, defaultValue int32) int32 {
	if v, ok := d.Props[name]; ok {
		return v
	}
	return defaultValue
}

// ResolveLevelData parses all level data from a level list directory.
func ResolveLevelData(levelList *wz.ImgDir) map[int32]*ItemOptionLevelData {
	result := make(map[int32]*ItemOptionLevelData)

	if levelList == nil {
		return result
	}

	for i := range levelList.ImgDirs {
		levelDir := &levelList.ImgDirs[i]
		level, err := strconv.ParseInt(levelDir.Name, 10, 32)
		if err != nil {
			continue
		}

		result[int32(level)] = NewItemOptionLevelData(int32(level), levelDir)
	}

	return result
}
