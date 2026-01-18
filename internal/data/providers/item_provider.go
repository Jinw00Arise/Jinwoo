package providers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/item"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

var ItemTypes = []string{
	"Consume", "Etc", "Install", "Cash",
}

var EquipTypes = []string{
	"Accessory", "Cap", "Cape", "Coat", "Dragon", "Face", "Glove", "Hair",
	"Longcoat", "Mechanic", "Pants", "PetEquip", "Ring", "Shield", "Shoes",
	"Taming", "Weapon",
}

type ItemProvider struct {
	wz *wz.WzProvider

	mu               sync.RWMutex
	itemInfos        map[int32]*item.ItemInfo
	itemRewardInfos  map[int32]*item.ItemRewardInfo
	mobSummonInfos   map[int32]*item.MobSummonInfo
	petActions       map[int32]map[int32]*item.PetInteraction
	specialItemNames map[int32]string
	itemOptionInfos  map[int32]*item.ItemOptionInfo
}

func NewItemProvider(wzProvider *wz.WzProvider) (*ItemProvider, error) {
	p := &ItemProvider{
		wz:               wzProvider,
		itemInfos:        make(map[int32]*item.ItemInfo),
		itemRewardInfos:  make(map[int32]*item.ItemRewardInfo),
		mobSummonInfos:   make(map[int32]*item.MobSummonInfo),
		petActions:       make(map[int32]map[int32]*item.PetInteraction),
		specialItemNames: make(map[int32]string),
		itemOptionInfos:  make(map[int32]*item.ItemOptionInfo),
	}

	if err := p.loadItemInfos(); err != nil {
		return nil, err
	}

	if err := p.loadItemNames(); err != nil {
		return nil, err
	}

	if err := p.loadItemOptionInfos(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *ItemProvider) loadItemInfos() error {
	itemDir := p.wz.Dir("Item.wz")

	// Load standard item types (Consume, Etc, Install, Cash)
	for _, typeName := range ItemTypes {
		typeDir := itemDir.Dir(typeName)
		images, err := typeDir.GetAllImages()
		if err != nil {
			return fmt.Errorf("could not resolve Item.wz/%s: %w", typeName, err)
		}

		for _, imgName := range images.Order {
			img := images.Items[imgName]
			root := img.Root()
			if root == nil {
				continue
			}

			// Each child of the image root is an item entry
			for i := range root.ImgDirs {
				itemEntry := &root.ImgDirs[i]
				itemID, err := strconv.ParseInt(itemEntry.Name, 10, 32)
				if err != nil {
					continue // skip non-numeric entries
				}

				id := int32(itemID)

				// Get info and spec directories
				infoDir := itemEntry.Get("info")
				specDir := itemEntry.Get("spec")

				p.itemInfos[id] = item.NewItemInfo(id, infoDir, specDir)

				// Item reward info
				if rewardDir := itemEntry.Get("reward"); rewardDir != nil {
					p.itemRewardInfos[id] = item.NewItemRewardInfo(id, rewardDir)
				}

				// Mob summon info
				if mobDir := itemEntry.Get("mob"); mobDir != nil {
					p.mobSummonInfos[id] = item.NewMobSummonInfo(id, mobDir)
				}
			}
		}
	}

	// Load pet items (different structure - each .img is one pet)
	petDir := itemDir.Dir("Pet")
	petImages, err := petDir.GetAllImages()
	if err != nil {
		return fmt.Errorf("could not resolve Item.wz/Pet: %w", err)
	}

	for _, imgName := range petImages.Order {
		img := petImages.Items[imgName]

		// Pet image name is the item ID (e.g., "5000000")
		itemID, err := strconv.ParseInt(imgName, 10, 32)
		if err != nil {
			continue
		}

		id := int32(itemID)
		root := img.Root()
		if root == nil {
			continue
		}

		// Get info directory directly from root
		infoDir := root.Get("info")
		p.itemInfos[id] = item.NewItemInfo(id, infoDir, nil)

		// Pet interactions
		interactDir := root.Get("interact")
		if interactDir == nil {
			continue
		}

		actions := make(map[int32]*item.PetInteraction)
		for j := range interactDir.ImgDirs {
			actionEntry := &interactDir.ImgDirs[j]
			action, err := strconv.ParseInt(actionEntry.Name, 10, 32)
			if err != nil {
				continue
			}

			actions[int32(action)] = item.NewPetInteraction(actionEntry)
		}

		if len(actions) > 0 {
			p.petActions[id] = actions
		}
	}

	return nil
}

func (p *ItemProvider) loadItemNames() error {
	itemDir := p.wz.Dir("Item.wz")
	specialDir := itemDir.Dir("Special")

	// Load from 0910.img and 0911.img
	for _, imgName := range []string{"0910", "0911"} {
		img, err := specialDir.Image(imgName)
		if err != nil {
			return fmt.Errorf("could not resolve Item.wz/Special/%s.img: %w", imgName, err)
		}

		if err := p.loadItemNamesFromImage(img); err != nil {
			return err
		}
	}

	return nil
}

func (p *ItemProvider) loadItemNamesFromImage(img *wz.WzImage) error {
	root := img.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		entry := &root.ImgDirs[i]
		itemID, err := strconv.ParseInt(entry.Name, 10, 32)
		if err != nil {
			continue
		}

		// Get the name string
		for _, strNode := range entry.Strings {
			if strNode.Name == "name" {
				p.specialItemNames[int32(itemID)] = strNode.Value
				break
			}
		}
	}

	return nil
}

func (p *ItemProvider) loadItemOptionInfos() error {
	// Item options are in Item.wz/ItemOption.img
	itemDir := p.wz.Dir("Item.wz")
	optionImg, err := itemDir.Image("ItemOption")
	if err != nil {
		return fmt.Errorf("could not resolve Item.wz/ItemOption.img: %w", err)
	}

	root := optionImg.Root()
	if root == nil {
		return nil
	}

	for i := range root.ImgDirs {
		optionEntry := &root.ImgDirs[i]
		optionID, err := strconv.ParseInt(optionEntry.Name, 10, 32)
		if err != nil {
			continue
		}

		p.itemOptionInfos[int32(optionID)] = item.NewItemOptionInfo(int32(optionID), optionEntry)
	}

	return nil
}

// GetItemInfo returns the item info for the given item ID.
func (p *ItemProvider) GetItemInfo(itemID int32) *item.ItemInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.itemInfos[itemID]
}

// GetItemRewardInfo returns the reward info for the given item ID.
func (p *ItemProvider) GetItemRewardInfo(itemID int32) *item.ItemRewardInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.itemRewardInfos[itemID]
}

// GetMobSummonInfo returns the mob summon info for the given item ID.
func (p *ItemProvider) GetMobSummonInfo(itemID int32) *item.MobSummonInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.mobSummonInfos[itemID]
}

// GetPetActions returns the pet interactions for the given pet item ID.
func (p *ItemProvider) GetPetActions(itemID int32) map[int32]*item.PetInteraction {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.petActions[itemID]
}

// HasItemInfo returns whether the item info exists.
func (p *ItemProvider) HasItemInfo(itemID int32) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.itemInfos[itemID]
	return ok
}

// GetSpecialItemName returns the special item name for the given item ID.
func (p *ItemProvider) GetSpecialItemName(itemID int32) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	name, ok := p.specialItemNames[itemID]
	return name, ok
}

// GetItemOptionInfo returns the item option info for the given option ID.
func (p *ItemProvider) GetItemOptionInfo(optionID int32) *item.ItemOptionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.itemOptionInfos[optionID]
}

// GetMatchingItemOptions returns all item options matching the given criteria.
func (p *ItemProvider) GetMatchingItemOptions(grade item.ItemGrade, reqLevel int32, bodyPart item.BodyPart) []*item.ItemOptionInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*item.ItemOptionInfo
	for _, optionInfo := range p.itemOptionInfos {
		if optionInfo.IsMatchingGrade(grade) &&
			optionInfo.IsMatchingLevel(reqLevel) &&
			optionInfo.IsMatchingType(bodyPart) {
			result = append(result, optionInfo)
		}
	}
	return result
}
