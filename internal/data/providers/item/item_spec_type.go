package item

import (
	"fmt"

	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
)

type ItemSpecsKey string

const (
	// HP/MP
	KeySpecHP     ItemSpecsKey = "hp"
	KeySpecMP     ItemSpecsKey = "mp"
	KeySpecHPR    ItemSpecsKey = "hpR"
	KeySpecMPR    ItemSpecsKey = "mpR"
	KeySpecTime   ItemSpecsKey = "time"
	KeySpecPad    ItemSpecsKey = "pad"
	KeySpecMad    ItemSpecsKey = "mad"
	KeySpecPdd    ItemSpecsKey = "pdd"
	KeySpecMdd    ItemSpecsKey = "mdd"
	KeySpecAcc    ItemSpecsKey = "acc"
	KeySpecEva    ItemSpecsKey = "eva"
	KeySpecSpeed  ItemSpecsKey = "speed"
	KeySpecJump   ItemSpecsKey = "jump"

	// Morph
	KeySpecMorph ItemSpecsKey = "morph"

	// Buff durations
	KeySpecBuffTime ItemSpecsKey = "buffTime"

	// Prob
	KeySpecProb ItemSpecsKey = "prob"

	// Consumable effects
	KeySpecThaw        ItemSpecsKey = "thaw"
	KeySpecCure        ItemSpecsKey = "cure"
	KeySpecMoveTo      ItemSpecsKey = "moveTo"
	KeySpecCP          ItemSpecsKey = "cp"
	KeySpecNuffSkill   ItemSpecsKey = "nuffSkill"
	KeySpecCooltime    ItemSpecsKey = "cooltime"

	// Experience/Meso
	KeySpecExp         ItemSpecsKey = "exp"
	KeySpecExpR        ItemSpecsKey = "expR"
	KeySpecMeso        ItemSpecsKey = "meso"
	KeySpecMesoR       ItemSpecsKey = "mesoR"

	// Item effects
	KeySpecItemUp      ItemSpecsKey = "itemUp"
	KeySpecItemUpByItem ItemSpecsKey = "itemUpByItem"
	KeySpecMesoupByItem ItemSpecsKey = "mesoupByItem"

	// Barrier
	KeySpecBarrier     ItemSpecsKey = "barrier"
	KeySpecBarrierType ItemSpecsKey = "barrierType"

	// Ignore
	KeySpecIgnoreDAM   ItemSpecsKey = "ignoreDam"
	KeySpecIgnoreDAMr  ItemSpecsKey = "ignoreDamR"

	// Party
	KeySpecParty       ItemSpecsKey = "party"

	// Pet
	KeySpecRepleteness ItemSpecsKey = "repleteness"
	KeySpecTameness    ItemSpecsKey = "tameness"

	// World Map
	KeySpecReturnMapQR ItemSpecsKey = "returnMapQR"
)

var itemSpecSchema = map[ItemSpecsKey]ItemValueKind{
	// HP/MP
	KeySpecHP:     ValueInt,
	KeySpecMP:     ValueInt,
	KeySpecHPR:    ValueInt,
	KeySpecMPR:    ValueInt,
	KeySpecTime:   ValueInt,
	KeySpecPad:    ValueInt,
	KeySpecMad:    ValueInt,
	KeySpecPdd:    ValueInt,
	KeySpecMdd:    ValueInt,
	KeySpecAcc:    ValueInt,
	KeySpecEva:    ValueInt,
	KeySpecSpeed:  ValueInt,
	KeySpecJump:   ValueInt,

	// Morph
	KeySpecMorph: ValueInt,

	// Buff durations
	KeySpecBuffTime: ValueInt,

	// Prob
	KeySpecProb: ValueInt,

	// Consumable effects
	KeySpecThaw:      ValueBool,
	KeySpecCure:      ValueString,
	KeySpecMoveTo:    ValueInt,
	KeySpecCP:        ValueInt,
	KeySpecNuffSkill: ValueInt,
	KeySpecCooltime:  ValueInt,

	// Experience/Meso
	KeySpecExp:   ValueInt,
	KeySpecExpR:  ValueInt,
	KeySpecMeso:  ValueInt,
	KeySpecMesoR: ValueInt,

	// Item effects
	KeySpecItemUp:       ValueBool,
	KeySpecItemUpByItem: ValueInt,
	KeySpecMesoupByItem: ValueInt,

	// Barrier
	KeySpecBarrier:     ValueInt,
	KeySpecBarrierType: ValueInt,

	// Ignore
	KeySpecIgnoreDAM:  ValueInt,
	KeySpecIgnoreDAMr: ValueInt,

	// Party
	KeySpecParty: ValueBool,

	// Pet
	KeySpecRepleteness: ValueInt,
	KeySpecTameness:    ValueInt,

	// World Map
	KeySpecReturnMapQR: ValueInt,
}

type ItemSpecs map[ItemSpecsKey]ItemValue

func (s ItemSpecs) Get(key ItemSpecsKey) (ItemValue, bool) {
	v, ok := s[key]
	return v, ok
}

func (s ItemSpecs) GetInt(key ItemSpecsKey) (int32, error) {
	v, ok := s[key]
	if !ok {
		return 0, fmt.Errorf("missing item spec key: %s", key)
	}
	if v.Kind != ValueInt {
		return 0, fmt.Errorf("item spec key %s is not int (got %d)", key, v.Kind)
	}
	return v.Int, nil
}

func (s ItemSpecs) GetBool(key ItemSpecsKey) (bool, error) {
	v, ok := s[key]
	if !ok {
		return false, fmt.Errorf("missing item spec key: %s", key)
	}
	if v.Kind != ValueBool {
		return false, fmt.Errorf("item spec key %s is not bool (got %d)", key, v.Kind)
	}
	return v.Bool, nil
}

func (s ItemSpecs) GetString(key ItemSpecsKey) (string, error) {
	v, ok := s[key]
	if !ok {
		return "", fmt.Errorf("missing item spec key: %s", key)
	}
	if v.Kind != ValueString {
		return "", fmt.Errorf("item spec key %s is not string (got %d)", key, v.Kind)
	}
	return v.String, nil
}

func (s ItemSpecs) GetDir(key ItemSpecsKey) (*wz.ImgDir, error) {
	v, ok := s[key]
	if !ok {
		return nil, fmt.Errorf("missing item spec key: %s", key)
	}
	if v.Kind != ValueDir {
		return nil, fmt.Errorf("item spec key %s is not dir (got %d)", key, v.Kind)
	}
	return v.Dir, nil
}

func (s ItemSpecs) Has(key ItemSpecsKey) bool {
	_, ok := s[key]
	return ok
}
