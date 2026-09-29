package db

type BackupEntity interface {
	GetID() int
	GetName() string
}

func (e View) GetID() int {
	return e.ID
}

func (e View) GetName() string {
	return e.Title
}

func (e Schedule) GetName() string {
	return e.Name
}

func (e Template) GetID() int {
	return e.ID
}

func (e Template) GetName() string {
	return e.Name
}

func (e Inventory) GetID() int {
	return e.ID
}

func (e Inventory) GetName() string {
	return e.Name
}

func (key AccessKey) GetID() int {
	return key.ID
}

func (key AccessKey) GetName() string {
	return key.Name
}

func (e Repository) GetID() int {
	return e.ID
}

func (e Repository) GetName() string {
	return e.Name
}

func (e Environment) GetID() int {
	return e.ID
}

func (e Environment) GetName() string {
	return e.Name
}

func (e SecretStorage) GetID() int {
	return e.ID
}

func (e SecretStorage) GetName() string {
	return e.Name
}

func (e Role) GetID() int {
	return e.ID
}

func (e Role) GetName() string {
	return e.Name
}

func (e TemplateVault) GetID() int {
	return e.ID
}

func (e Task) GetID() int {
	return e.ID
}

func (e Integration) GetID() int {
	return e.ID
}

func (e Runner) GetID() int {
	return e.ID
}

func (e Runner) GetName() string {
	return e.Name
}

func (e Project) GetID() int {
	return e.ID
}
func (e User) GetID() int {
	return e.ID
}
