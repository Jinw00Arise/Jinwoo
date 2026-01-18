package item

import "github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"

type PetInteraction struct {
	Prob       int32
	Inc        int32
	LevelLimit int32
}

func NewPetInteraction(interactDir *wz.ImgDir) *PetInteraction {
	interaction := &PetInteraction{}

	if interactDir == nil {
		return interaction
	}

	if prob, err := interactDir.GetInt("prob"); err == nil {
		interaction.Prob = prob
	}
	if inc, err := interactDir.GetInt("inc"); err == nil {
		interaction.Inc = inc
	}
	if levelLimit, err := interactDir.GetInt("l0"); err == nil {
		interaction.LevelLimit = levelLimit
	}

	return interaction
}

func (p *PetInteraction) GetProb() int32 {
	return p.Prob
}

func (p *PetInteraction) GetInc() int32 {
	return p.Inc
}

func (p *PetInteraction) GetLevelLimit() int32 {
	return p.LevelLimit
}
