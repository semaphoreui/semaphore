package export

import (
	"fmt"

	"github.com/semaphoreui/semaphore/db"
)

type UserExporter struct {
	ValueMap[db.User]
	MergeExisting bool
}

func (a *UserExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	users, err := store.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return err
	}

	return a.appendValues(users, GlobalScope)
}

func (a *UserExporter) restore(store db.Store, exporter DataExporter, progress Progress) error {
	for _, val := range a.values {
		old := val.value

		obj, err := store.ImportUser(db.UserWithPwd{Pwd: old.Password, User: old})
		if err != nil {
			if !a.MergeExisting {
				return err
			}

			existing, lookupErr := store.GetUserByLoginOrEmail(old.Username, old.Email)
			if lookupErr != nil {
				return fmt.Errorf("import user failed: %w; lookup existing user: %w", err, lookupErr)
			}

			obj = existing
		}

		err = exporter.mapIntKeys(a.getName(), GlobalScope, old.ID, obj.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *UserExporter) getName() string {
	return User
}
