package item

import "github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"

type MobSummon struct {
	MobID int32
	Prob  int32
}

type MobSummonInfo struct {
	itemID  int32
	summons []MobSummon
}

func NewMobSummonInfo(itemID int32, mobDir *wz.ImgDir) *MobSummonInfo {
	info := &MobSummonInfo{
		itemID:  itemID,
		summons: make([]MobSummon, 0),
	}

	if mobDir == nil {
		return info
	}

	for i := range mobDir.ImgDirs {
		mobEntry := &mobDir.ImgDirs[i]
		summon := MobSummon{}

		if id, err := mobEntry.GetInt("id"); err == nil {
			summon.MobID = id
		}
		if prob, err := mobEntry.GetInt("prob"); err == nil {
			summon.Prob = prob
		}

		info.summons = append(info.summons, summon)
	}

	return info
}

func (m *MobSummonInfo) GetItemID() int32 {
	return m.itemID
}

func (m *MobSummonInfo) GetSummons() []MobSummon {
	return m.summons
}
