package sql

import (
	"github.com/go-gorp/gorp/v3"
	"strings"
)

type migration_2_8_26 struct {
	db *SqlDb
}

func (m migration_2_8_26) PostApply(tx *gorp.Transaction) error {
	rows, err := tx.Query(m.db.PrepareQuery("SELECT id, git_url FROM project__repository"))
	if err != nil {
		return err
	}

	repoUrls := make(map[string]string)

	for rows.Next() {
		var id, url string

		err3 := rows.Scan(&id, &url)
		if err3 != nil {
			continue
		}

		repoUrls[id] = url
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	for id, url := range repoUrls {
		branch := "master"
		parts := strings.Split(url, "#")
		if len(parts) > 1 {
			url, branch = parts[0], parts[1]
		}
		q := m.db.PrepareQuery("UPDATE project__repository " +
			"SET git_url = ?, git_branch = ? " +
			"WHERE id = ?")
		_, err = tx.Exec(q, url, branch, id)

		if err != nil {
			return err
		}
	}

	return nil
}
