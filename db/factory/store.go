package factory

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"

	//"github.com/ansible-semaphore/semaphore/db/sql"
)

func CreateStore() db.Store {
	return &bolt.BoltDb{}
	//return &sql.SqlDb{}
}
