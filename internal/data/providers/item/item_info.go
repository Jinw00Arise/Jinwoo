package item

import "github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"

type ItemInfo struct {
	itemID    int32
	itemInfos ItemInfos
	itemSpecs ItemSpecs
}

func NewItemInfo(itemID int32, infoDir *wz.ImgDir, specDir *wz.ImgDir) *ItemInfo {
	info := &ItemInfo{
		itemID:    itemID,
		itemInfos: make(ItemInfos),
		itemSpecs: make(ItemSpecs),
	}

	if infoDir != nil {
		info.itemInfos = parseItemInfos(infoDir)
	}

	if specDir != nil {
		info.itemSpecs = parseItemSpecs(specDir)
	}

	return info
}

func (i *ItemInfo) GetItemID() int32 {
	return i.itemID
}

func (i *ItemInfo) GetItemInfos() ItemInfos {
	return i.itemInfos
}

func (i *ItemInfo) GetItemSpecs() ItemSpecs {
	return i.itemSpecs
}

// GetInfo returns the int value for an info key, or an error if missing/wrong type.
func (i *ItemInfo) GetInfo(key ItemInfosKey) (int32, error) {
	return i.itemInfos.GetInt(key)
}

// GetInfoOr returns the int value for an info key, or the default if missing.
func (i *ItemInfo) GetInfoOr(key ItemInfosKey, defaultValue int32) int32 {
	if v, ok := i.itemInfos.Get(key); ok && v.Kind == ValueInt {
		return v.Int
	}
	return defaultValue
}

// GetSpec returns the int value for a spec key, or an error if missing/wrong type.
func (i *ItemInfo) GetSpec(key ItemSpecsKey) (int32, error) {
	return i.itemSpecs.GetInt(key)
}

// GetSpecOr returns the int value for a spec key, or the default if missing.
func (i *ItemInfo) GetSpecOr(key ItemSpecsKey, defaultValue int32) int32 {
	if v, ok := i.itemSpecs.Get(key); ok && v.Kind == ValueInt {
		return v.Int
	}
	return defaultValue
}

// HasInfo returns whether the info key exists.
func (i *ItemInfo) HasInfo(key ItemInfosKey) bool {
	return i.itemInfos.Has(key)
}

// HasSpec returns whether the spec key exists.
func (i *ItemInfo) HasSpec(key ItemSpecsKey) bool {
	return i.itemSpecs.Has(key)
}

func parseItemInfos(dir *wz.ImgDir) ItemInfos {
	out := make(ItemInfos)

	// Parse int nodes (handles both ValueInt and ValueBool)
	for _, intNode := range dir.Ints {
		key := ItemInfosKey(intNode.Name)
		kind, ok := itemInfoSchema[key]
		if !ok {
			continue
		}
		switch kind {
		case ValueInt:
			out[key] = ItemValue{Kind: ValueInt, Int: intNode.Value}
		case ValueBool:
			out[key] = ItemValue{Kind: ValueBool, Bool: intNode.Value == 1}
		}
	}

	// Parse string nodes
	for _, strNode := range dir.Strings {
		key := ItemInfosKey(strNode.Name)
		if _, ok := itemInfoSchema[key]; ok {
			out[key] = ItemValue{Kind: ValueString, String: strNode.Value}
		}
	}

	// Parse child directories (ValueDir)
	for i := range dir.ImgDirs {
		key := ItemInfosKey(dir.ImgDirs[i].Name)
		if kind, ok := itemInfoSchema[key]; ok && kind == ValueDir {
			out[key] = ItemValue{Kind: ValueDir, Dir: &dir.ImgDirs[i]}
		}
	}

	return out
}

func parseItemSpecs(dir *wz.ImgDir) ItemSpecs {
	out := make(ItemSpecs)

	// Parse int nodes (handles both ValueInt and ValueBool)
	for _, intNode := range dir.Ints {
		key := ItemSpecsKey(intNode.Name)
		kind, ok := itemSpecSchema[key]
		if !ok {
			continue
		}
		switch kind {
		case ValueInt:
			out[key] = ItemValue{Kind: ValueInt, Int: intNode.Value}
		case ValueBool:
			out[key] = ItemValue{Kind: ValueBool, Bool: intNode.Value == 1}
		}
	}

	// Parse string nodes
	for _, strNode := range dir.Strings {
		key := ItemSpecsKey(strNode.Name)
		if _, ok := itemSpecSchema[key]; ok {
			out[key] = ItemValue{Kind: ValueString, String: strNode.Value}
		}
	}

	// Parse child directories (ValueDir)
	for i := range dir.ImgDirs {
		key := ItemSpecsKey(dir.ImgDirs[i].Name)
		if kind, ok := itemSpecSchema[key]; ok && kind == ValueDir {
			out[key] = ItemValue{Kind: ValueDir, Dir: &dir.ImgDirs[i]}
		}
	}

	return out
}
