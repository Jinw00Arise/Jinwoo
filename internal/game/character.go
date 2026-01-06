package game

import (
	"github.com/Jinw00Arise/Jinwoo/internal/database/models"
)

// CharacterWrapper wraps a database Character model to implement the Character interface.
type CharacterWrapper struct {
	model *models.Character
}

// WrapCharacter wraps a database model to implement the Character interface.
func WrapCharacter(model *models.Character) *CharacterWrapper {
	return &CharacterWrapper{model: model}
}

// Model returns the underlying database model.
func (c *CharacterWrapper) Model() *models.Character {
	return c.model
}

// Identity
func (c *CharacterWrapper) GetID() uint          { return c.model.ID }
func (c *CharacterWrapper) GetAccountID() uint   { return c.model.AccountID }
func (c *CharacterWrapper) GetName() string      { return c.model.Name }
func (c *CharacterWrapper) GetGender() byte      { return c.model.Gender }

// Appearance
func (c *CharacterWrapper) GetSkinColor() byte { return c.model.SkinColor }
func (c *CharacterWrapper) GetFace() int32     { return c.model.Face }
func (c *CharacterWrapper) GetHair() int32     { return c.model.Hair }

// Stats
func (c *CharacterWrapper) GetLevel() byte  { return c.model.Level }
func (c *CharacterWrapper) GetJob() int16   { return c.model.Job }
func (c *CharacterWrapper) GetSTR() int16   { return c.model.STR }
func (c *CharacterWrapper) GetDEX() int16   { return c.model.DEX }
func (c *CharacterWrapper) GetINT() int16   { return c.model.INT }
func (c *CharacterWrapper) GetLUK() int16   { return c.model.LUK }
func (c *CharacterWrapper) GetHP() int32    { return c.model.HP }
func (c *CharacterWrapper) GetMaxHP() int32 { return c.model.MaxHP }
func (c *CharacterWrapper) GetMP() int32    { return c.model.MP }
func (c *CharacterWrapper) GetMaxMP() int32 { return c.model.MaxMP }
func (c *CharacterWrapper) GetAP() int16    { return c.model.AP }
func (c *CharacterWrapper) GetSP() int16    { return c.model.SP }
func (c *CharacterWrapper) GetEXP() int32   { return c.model.EXP }
func (c *CharacterWrapper) GetFame() int16  { return c.model.Fame }
func (c *CharacterWrapper) GetMeso() int32  { return c.model.Meso }

// Location
func (c *CharacterWrapper) GetMapID() int32      { return c.model.MapID }
func (c *CharacterWrapper) GetSpawnPoint() byte  { return c.model.SpawnPoint }

// Setters
func (c *CharacterWrapper) SetHP(v int32)         { c.model.HP = v }
func (c *CharacterWrapper) SetMP(v int32)         { c.model.MP = v }
func (c *CharacterWrapper) SetEXP(v int32)        { c.model.EXP = v }
func (c *CharacterWrapper) SetMeso(v int32)       { c.model.Meso = v }
func (c *CharacterWrapper) SetFame(v int16)       { c.model.Fame = v }
func (c *CharacterWrapper) SetMapID(v int32)      { c.model.MapID = v }
func (c *CharacterWrapper) SetSpawnPoint(v byte)  { c.model.SpawnPoint = v }
func (c *CharacterWrapper) SetLevel(v byte)       { c.model.Level = v }

