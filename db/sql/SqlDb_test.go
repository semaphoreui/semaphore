package sql

import (
	"testing"

	"github.com/go-gorp/gorp/v3"
)

func TestSqlDbConnection_CloseWithoutConnect(t *testing.T) {
	var conn SqlDbConnection
	conn.Close() // must not panic when never connected
}

func TestSqlDb_CloseWithoutConnect(t *testing.T) {
	store := CreateDb("sqlite")
	store.Close() // must not panic when never connected
}

func TestValidatePort(t *testing.T) {
	d := SqlDb{}
	q := d.connection.prepareQueryWithDialect("select * from `test` where id = ?, email = ?", gorp.PostgresDialect{})
	if q != "select * from \"test\" where id = $1, email = $2" {
		t.Error("invalid postgres query")
	}
}
