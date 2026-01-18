package item

import "github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"

type ItemReward struct {
	ItemID int32
	Count  int32
	Prob   int32
}

type ItemRewardInfo struct {
	itemID  int32
	rewards []ItemReward
}

func NewItemRewardInfo(itemID int32, rewardDir *wz.ImgDir) *ItemRewardInfo {
	info := &ItemRewardInfo{
		itemID:  itemID,
		rewards: make([]ItemReward, 0),
	}

	if rewardDir == nil {
		return info
	}

	for i := range rewardDir.ImgDirs {
		rewardEntry := &rewardDir.ImgDirs[i]
		reward := ItemReward{}

		if id, err := rewardEntry.GetInt("item"); err == nil {
			reward.ItemID = id
		}
		if count, err := rewardEntry.GetInt("count"); err == nil {
			reward.Count = count
		}
		if prob, err := rewardEntry.GetInt("prob"); err == nil {
			reward.Prob = prob
		}

		info.rewards = append(info.rewards, reward)
	}

	return info
}

func (r *ItemRewardInfo) GetItemID() int32 {
	return r.itemID
}

func (r *ItemRewardInfo) GetRewards() []ItemReward {
	return r.rewards
}
