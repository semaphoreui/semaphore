package db

type BackupEntity interface {
	GetID() int
	GetName() string
}

type BackupSluggedEntity interface {
	GetSlug() string
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

func (e SecretStorage) GetID() int {
	return e.ID
}

func (e SecretStorage) GetName() string {
	return e.Name
}

func (e Role) GetID() int {
	panic("Role does not implement GetID")
}

func (e Role) GetSlug() string {
	return e.Slug
}

func (e Role) GetName() string {
	if e.ProjectID == nil {
		return e.Slug
	}
	return e.Name
}
