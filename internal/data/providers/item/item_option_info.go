package item

import (
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

type ItemOptionInfo struct {
	itemOptionID int32
	reqLevel     int32
	optionType   ItemOptionType
	levelData    map[int32]*ItemOptionLevelData
}

func NewItemOptionInfo(itemOptionID int32, optionDir *wz.ImgDir) *ItemOptionInfo {
	info := &ItemOptionInfo{
		itemOptionID: itemOptionID,
		reqLevel:     0,
		optionType:   ItemOptionAnyEquip,
		levelData:    make(map[int32]*ItemOptionLevelData),
	}

	if optionDir == nil {
		return info
	}

	// Parse info
	if infoDir := optionDir.Get("info"); infoDir != nil {
		if reqLevel, err := infoDir.GetInt("reqLevel"); err == nil {
			info.reqLevel = reqLevel
		}
		if optionType, err := infoDir.GetInt("optionType"); err == nil {
			info.optionType = ItemOptionTypeFromValue(optionType)
		}
	}

	// Parse level data
	if levelList := optionDir.Get("level"); levelList != nil {
		info.levelData = ResolveLevelData(levelList)
	}

	return info
}

func (i *ItemOptionInfo) GetItemOptionID() int32 {
	return i.itemOptionID
}

func (i *ItemOptionInfo) GetReqLevel() int32 {
	return i.reqLevel
}

func (i *ItemOptionInfo) GetOptionType() ItemOptionType {
	return i.optionType
}

func (i *ItemOptionInfo) GetLevelData(optionLevel int32) (*ItemOptionLevelData, bool) {
	data, ok := i.levelData[optionLevel]
	return data, ok
}

func (i *ItemOptionInfo) IsMatchingGrade(grade ItemGrade) bool {
	return grade.MatchesOptionID(i.itemOptionID)
}

func (i *ItemOptionInfo) IsMatchingLevel(itemReqLevel int32) bool {
	return i.reqLevel <= itemReqLevel
}

func (i *ItemOptionInfo) IsMatchingType(bodyPart BodyPart) bool {
	return bodyPart.MatchesOptionType(i.optionType)
}
