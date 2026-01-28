package db_migration

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	projectService "github.com/semaphoreui/semaphore/services/project"
)

type Migrator struct {
	oldStore db.Store
	newStore db.Store

	userIDs map[int]db.User
}

func Migrate(oldStore, newStore db.Store) error {
	migrator := &Migrator{}
	return migrator.Migrate(oldStore, newStore)
}

func (m *Migrator) Migrate(oldStore, newStore db.Store) error {
	m.oldStore = oldStore
	m.newStore = newStore

	m.userIDs = make(map[int]db.User)

	fmt.Println("Migrating users...")
	if err := m.migrateUsers(); err != nil {
		return err
	}

	fmt.Println("Migrating projects...")
	if err := m.migrateProjects(); err != nil {
		return err
	}

	fmt.Println("Migration completed successfully.")
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

func (m *Migrator) migrateProjects() error {
	users, err := m.oldStore.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	migrated := map[int]bool{}
	for _, user := range users {
		newUser := m.userIDs[user.ID]
		projects, err := m.oldStore.GetProjects(user.ID)

		if err != nil {
			return err
		}

		for _, project := range projects {
			if !migrated[project.ID] {
				err = m.migrateProject(&newUser, &project)
				if err != nil {
					return err
				}

				migrated[project.ID] = true
			}
		}
	}
	return nil
}

func (m *Migrator) migrateProject(user *db.User, project *db.Project) error {
	backup, err := projectService.GetBackup(project.ID, m.oldStore)

	if err != nil {
		return err
	}

	if err := backup.Verify(); err != nil {
		return err
	}

	_, err = backup.Restore(*user, m.newStore)
	return err
}
