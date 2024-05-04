package gormdb

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/glebarez/sqlite"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GormdbStore implements the Store interface.
type GormdbStore struct {
	driver          string
	username        string
	password        string
	hostname        string
	database        string
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	params          url.Values

	handle *gorm.DB
}

// Info returns some basic db informations.
func (s *GormdbStore) Info() map[string]interface{} {
	result := make(map[string]interface{})
	result["driver"] = s.driver
	result["database"] = s.database

	if s.hostname != "" {
		result["hostname"] = s.hostname
	}

	if s.username != "" {
		result["username"] = s.username
	}

	return result
}

// Prepare is preparing some database behavior.
func (s *GormdbStore) Prepare() error {
	sqldb, err := s.handle.DB()

	if err != nil {
		return err
	}

	switch s.driver {
	case "mysql", "mariadb":
		sqldb.SetMaxOpenConns(s.maxOpenConns)
		sqldb.SetMaxIdleConns(s.maxIdleConns)
		sqldb.SetConnMaxLifetime(s.connMaxLifetime)
	case "postgres", "postgresql":
		sqldb.SetMaxOpenConns(s.maxOpenConns)
		sqldb.SetMaxIdleConns(s.maxIdleConns)
		sqldb.SetConnMaxLifetime(s.connMaxLifetime)
	}

	return nil
}

// Open simply opens the database connection.
func (s *GormdbStore) Open() error {
	dialect, err := s.open()

	if err != nil {
		return err
	}

	handle, err := gorm.Open(
		dialect,
		&gorm.Config{
			Logger:               NewLogger(),
			DisableAutomaticPing: true,
		},
	)

	if err != nil {
		return err
	}

	s.handle = handle
	return s.Prepare()
}

// Close simply closes the database connection.
func (s *GormdbStore) Close() error {
	sqldb, err := s.handle.DB()

	if err != nil {
		return err
	}

	return sqldb.Close()
}

// Ping just tests the database connection.
func (s *GormdbStore) Ping() error {
	sqldb, err := s.handle.DB()

	if err != nil {
		return err
	}

	return sqldb.Ping()
}

// Migrate executes required db migrations.
func (s *GormdbStore) Migrate() error {
	migrate := gormigrate.New(
		s.handle,
		gormigrate.DefaultOptions,
		Migrations,
	)

	return migrate.Migrate()
}

// Admin creates an initial admin user within the database.
func (db *GormdbStore) Admin(username, password, email string) error {
	admin := &model.User{}

	if err := db.handle.Where(
		&model.User{
			Username: username,
		},
	).First(
		admin,
	).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	admin.Username = username
	admin.Password = password
	admin.Email = email
	admin.Admin = true

	if admin.Name == "" {
		admin.Name = "Admin"
	}

	tx := db.handle.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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

func (db *GormdbStore) open() (gorm.Dialector, error) {
	switch db.driver {
	case "sqlite", "sqlite3":
		dsn := fmt.Sprintf(
			"%s?%s",
			db.database,
			db.params.Encode(),
		)

		return sqlite.Open(dsn), nil
	case "mysql", "mariadb":
		dsn := fmt.Sprintf(
			"%s@(%s)/%s?%s",
			db.username,
			db.hostname,
			db.database,
			db.params.Encode(),
		)

		if db.password != "" {
			dsn = fmt.Sprintf(
				"%s:%s@(%s)/%s?%s",
				db.username,
				db.password,
				db.hostname,
				db.database,
				db.params.Encode(),
			)
		}

		return mysql.Open(dsn), nil
	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"%s@%s/%s?%s",
			db.username,
			db.hostname,
			db.database,
			db.params.Encode(),
		)

		if db.password != "" {
			dsn = fmt.Sprintf(
				"%s:%s@%s/%s?%s",
				db.username,
				db.password,
				db.hostname,
				db.database,
				db.params.Encode(),
			)
		}

		return postgres.Open(dsn), nil
	}

	return nil, nil
}

func (db *GormdbStore) sortQuery(query *gorm.DB, props db.ObjectProps, params model.RetrieveQueryParams) *gorm.DB {
	orderDirection := "ASC"
	if params.SortInverted {
		orderDirection = "DESC"
	}

	orderColumn := props.DefaultSortingColumn
	if slices.Contains(props.SortableColumns, params.SortBy) {
		orderColumn = params.SortBy
	}

	if orderColumn != "" {
		query = query.Order(strings.Join(
			[]string{orderColumn, orderDirection},
			" ",
		))
	}

	if params.Count > 0 {
		query = query.Limit(params.Count)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	return query
}

// NewStore initializes a new gorm store instance
func NewStore(cfg util.DbConfig) (store.Store, error) {
	client := &GormdbStore{
		driver:   cfg.Dialect,
		database: cfg.DbName,

		username: cfg.Username,
		password: cfg.Password,

		params: url.Values{},
	}

	if val, ok := cfg.Options["maxOpenConns"]; ok {
		cur, err := strconv.Atoi(
			val,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse maxOpenConns: %w", err)
		}

		client.maxOpenConns = cur
		delete(cfg.Options, "maxOpenConns")
	} else {
		client.maxOpenConns = 25
	}

	if val, ok := cfg.Options["maxIdleConns"]; ok {
		cur, err := strconv.Atoi(
			val,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse maxIdleConns: %w", err)
		}

		client.maxIdleConns = cur
		delete(cfg.Options, "maxIdelConns")
	} else {
		client.maxIdleConns = 25
	}

	if val, ok := cfg.Options["connMaxLifetime"]; ok {
		cur, err := time.ParseDuration(
			val,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to parse connMaxLifetime: %w", err)
		}

		client.connMaxLifetime = cur
		delete(cfg.Options, "connMaxLifetime")
	} else {
		client.connMaxLifetime = 5 * time.Minute
	}

	for key, val := range cfg.Options {
		client.params.Set(key, val)
	}

	switch client.driver {
	case "sqlite", "sqlite3":
		client.driver = "sqlite"
		client.database = cfg.Hostname

		client.params.Add("_pragma", "journal_mode(WAL)")
		client.params.Add("_pragma", "busy_timeout(5000)")
		client.params.Add("_pragma", "foreign_keys(1)")
	case "mysql", "mariadb":
		client.driver = "mysql"
		client.hostname = cfg.Hostname

		if _, ok := cfg.Options["charset"]; !ok {
			client.params.Set("charset", "utf8")
		}

		if _, ok := cfg.Options["parseTime"]; !ok {
			client.params.Set("parseTime", "True")
		}

		if _, ok := cfg.Options["loc"]; !ok {
			client.params.Set("loc", "Local")
		}
	case "postgres", "postgresql":
		client.driver = "postgres"
		client.hostname = cfg.Hostname

		if _, ok := cfg.Options["sslmode"]; !ok {
			client.params.Set("sslmode", "disable")
		}
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
