package factory

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/sql"
)

func CreateStore() db.Store {
	return &sql.SqlDb{}
}
