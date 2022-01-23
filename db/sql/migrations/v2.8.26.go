package migrations

import (
	"github.com/go-gorp/gorp/v3"
	"strings"
)

type Migration_2_8_26 struct {
	Sql *gorp.DbMap
}

func (m Migration_2_8_26) Apply() error {
	rows, err := m.Sql.Query("SELECT id, git_url FROM project__repository")
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var id, url string

		err3 := rows.Scan(&id, &url)
		if err3 != nil {
			continue
		}

		branch := "master"
		parts := strings.Split(url, "#")
		if len(parts) > 1 {
			url, branch = parts[0], parts[1]
		}
		_, _ = m.Sql.Exec("UPDATE project__repository "+
			"SET git_url = ?, git_branch = ? "+
			"WHERE id = ?", url, branch, id)
	}

	return nil
}
