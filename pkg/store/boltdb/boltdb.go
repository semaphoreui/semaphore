package boltdb

import (
	"context"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	bolt "go.etcd.io/bbolt"
)

// BoltdbStore implements the Store interface.
type BoltdbStore struct {
	database string
	perms    os.FileMode
	timeout  time.Duration

	handle *storm.DB
}

// Info returns some basic db informations.
func (db *BoltdbStore) Info() map[string]interface{} {
	result := make(map[string]interface{})
	result["driver"] = "boltdb"
	result["database"] = db.database
	result["perms"] = db.perms.String()
	result["timeout"] = db.timeout.String()

	return result
}

// Prepare is preparing some database behavior.
func (db *BoltdbStore) Prepare() error {
	return nil
}

// Open simply opens the database connection.
func (db *BoltdbStore) Open() error {
	handle, err := storm.Open(
		db.database,
		storm.BoltOptions(
			db.perms,
			&bolt.Options{
				Timeout: db.timeout,
			},
		),
	)

	if err != nil {
		return err
	}

	db.handle = handle
	return db.Prepare()
}

// Close simply closes the database connection.
func (db *BoltdbStore) Close() error {
	return db.handle.Close()
}

// Ping just tests the database connection.
func (db *BoltdbStore) Ping() error {
	return nil
}

// Migrate executes required db migrations.
func (db *BoltdbStore) Migrate() error {
	return nil
}

// Admin creates an initial admin user within the database.
func (db *BoltdbStore) Admin(username, password, email string) error {
	admin := &model.User{}

	if err := db.handle.Select(
		q.Eq("username", username),
	).First(admin); err != nil && err != storm.ErrNotFound {
		return err
	}

	admin.Username = username
	admin.Password = password
	admin.Email = email
	admin.Admin = true

	if admin.Name == "" {
		admin.Name = "Admin"
	}

	if admin.ID == 0 {
		if err := db.CreateUser(
			context.Background(),
			admin,
		); err != nil {
			return err
		}
	} else {
		if err := db.UpdateUser(
			context.Background(),
			admin,
		); err != nil {
			return err
		}
	}

	return nil
}

func (db *BoltdbStore) sortQuery(query storm.Query, props db.ObjectProps, params model.RetrieveQueryParams) storm.Query {
	if params.SortInverted {
		query = query.Reverse()
	}

	orderColumn := props.DefaultSortingColumn
	if slices.Contains(props.SortableColumns, params.SortBy) {
		orderColumn = params.SortBy
	}

	if orderColumn != "" {
		query = query.OrderBy(orderColumn)
	}

	if params.Count > 0 {
		query = query.Limit(params.Count)
	}

	if params.Offset > 0 {
		query = query.Skip(params.Offset)
	}

	return query
}

// NewStore initializes a new gorm store instance
func NewStore(cfg util.DbConfig) (store.Store, error) {
	client := &BoltdbStore{
		database: cfg.Hostname,
	}

	if val, ok := cfg.Options["perms"]; ok {
		res, err := strconv.ParseUint(val, 8, 32)

		if err != nil {
			client.perms = os.FileMode(0600)
		} else {
			client.perms = os.FileMode(res)
		}
	} else {
		client.perms = os.FileMode(0600)
	}

	if val, ok := cfg.Options["timeout"]; ok {
		res, err := time.ParseDuration(val)

		if err != nil {
			client.timeout = 1 * time.Second
		} else {
			client.timeout = res
		}
	} else {
		client.timeout = 1 * time.Second
	}

	return client, nil
}

// MustStore simply calls NewStore and panics on an error.
func MustStore(cfg util.DbConfig) store.Store {
	s, err := NewStore(cfg)

	if err != nil {
		panic(err)
	}

	return s
}
