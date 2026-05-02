package export

import (
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

	var userMap = make(map[string]*db.User)
	if a.MergeExisting {
		users, err := store.GetUsers(db.RetrieveQueryParams{})
		if err != nil {
			return err
		}
		for _, user := range users {
			userMap[user.Username] = &user
		}
	}

	for _, val := range a.values {
		var err error
		old := val.value
		var obj db.User

		if u, ok := userMap[old.Username]; ok && a.MergeExisting {
			obj = *u
		} else {
			obj, err = store.ImportUser(db.UserWithPwd{Pwd: old.Password, User: old})
			if err != nil {
				return err
			}
		}

		err = exporter.mapKeys(a.getName(), GlobalScope, old.GetDbKey(), obj.GetDbKey())
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *UserExporter) getName() string {
	return User
}
