package sql

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
	"testing"
)

var pool *mockDriver;

type sqlmock struct {
	ordered      bool
	dsn          string
	opened       int
	drv          *mockDriver
	converter    driver.ValueConverter
	queryMatcher QueryMatcher
	monitorPings bool

	expected []expectation
}

func init() {
	pool = &mockDriver{
		conns: make(map[string]*sqlmock),
	}
	SqlDb.sql.Register("sqlmock", pool)
}

func TestValidatePort(t *testing.T) {
	d := SqlDb{}
	q := d.prepareQueryWithDialect("select * from `test` where id = ?, email = ?", gorp.PostgresDialect{})
	if q != "select * from \"test\" where id = $1, email = $2" {
		t.Error("invalid postgres query")
	}
}

func TestGetAllObjects(t *testing.T) {

}
