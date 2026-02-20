package sql

import (
	"testing"

	"github.com/go-gorp/gorp/v3"
)

func TestValidatePort(t *testing.T) {
	d := SqlDb{}
	q := d.connection.prepareQueryWithDialect("select * from `test` where id = ?, email = ?", gorp.PostgresDialect{})
	if q != "select * from \"test\" where id = $1, email = $2" {
		t.Error("invalid postgres query")
	}
}
