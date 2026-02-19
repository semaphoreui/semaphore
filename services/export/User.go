package export

import "github.com/semaphoreui/semaphore/db"

type UserExporter struct {
	ValueMap[db.User]
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
			return err
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
