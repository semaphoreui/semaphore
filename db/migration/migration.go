package migration

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/export"
)

type Migrator struct {
	oldStore db.Store
	newStore db.Store

	userIDs map[int]db.User
}

func Migrate(oldStore, newStore db.Store, errLogSize int) error {
	migrator := &Migrator{}
	return migrator.Migrate(oldStore, newStore, errLogSize)
}

func (m *Migrator) Migrate(oldStore, newStore db.Store, errLogSize int) error {
	m.oldStore = oldStore
	m.newStore = newStore

	m.userIDs = make(map[int]db.User)

	if err := m.migrateProject(errLogSize); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) migrateUsers() error {
	users, err := m.oldStore.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	for _, user := range users {
		oldID := user.ID
		user.ID = 0
		newUser, err := m.newStore.ImportUser(db.UserWithPwd{Pwd: user.Password, User: user})
		if err != nil {
			return err
		}
		m.userIDs[oldID] = newUser
	}
	return nil
}

func (m *Migrator) migrateProject(errLogSize int) error {

	mapper := &export.TypeKeyMapper{Keys: make(map[string]map[string]map[export.EntityKey]export.EntityKey), IgnoreKeyNotFoundErr: true}
	p := export.InitProjectExporters(mapper)

	err := p.Load(m.oldStore)
	if err != nil {
		return err
	}

	err = p.Restore(m.newStore, errLogSize)
	if err != nil {
		return err
	}

	return err
}
