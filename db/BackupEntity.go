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

func (e AccessKey) GetID() int {
	return e.ID
}

func (e AccessKey) GetName() string {
	return e.Name
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
