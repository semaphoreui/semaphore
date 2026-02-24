package db

import (
	"strconv"
)

func NewKeyFromInt(key int) string {
	return strconv.Itoa(key)
}

func (e TemplateVault) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Task) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Integration) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Project) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}
func (e User) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}
func (e Template) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Environment) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Repository) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e SecretStorage) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (key AccessKey) GetDbKey() string {
	return NewKeyFromInt(key.ID)
}

func (e Inventory) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Role) GetDbKey() string {
	return e.Slug
}

func (e View) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e IntegrationAlias) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e IntegrationExtractValue) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e IntegrationMatcher) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Schedule) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e TaskStage) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e TemplateRolePerm) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e TaskParams) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e ProjectUser) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e TaskOutput) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e TaskStageResult) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Option) GetDbKey() string {
	return e.Key
}

func (e Event) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}

func (e Runner) GetDbKey() string {
	return NewKeyFromInt(e.ID)
}
