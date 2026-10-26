package sql

import (
	"strings"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
)

type migration_2_20_18 struct {
	db *SqlDb
}

// PostApply links every existing Variable Group secret to its key in project__environment_secret.
// Until now a group's secrets were "all access keys with its environment_id"; the type came from the
// key owner and the name from the key name, with a legacy `var.` / `env.` prefix on old rows. The
// key rows are left as they are: owner and environment_id keep marking the key as owned by the group.
func (m migration_2_20_18) PostApply(tx *gorp.Transaction) error {
	var keys []struct {
		ID            int    `db:"id"`
		ProjectID     int    `db:"project_id"`
		EnvironmentID int    `db:"environment_id"`
		Owner         string `db:"owner"`
		Name          string `db:"name"`
	}
	_, err := tx.Select(&keys, m.db.PrepareQuery(`
		select
		  k.id, e.project_id, k.environment_id, k.owner, k.name
		from access_key k
		join project__environment e on e.id = k.environment_id
		where k.environment_id is not null
		  and k.owner in ('variable', 'environment')
		order by k.id
	`))
	if err != nil {
		return err
	}

	type link struct {
		keyID     int
		projectID int
		envID     int
		secretTyp db.EnvironmentSecretType
		name      string
	}
	type slot struct {
		envID int
		typ   db.EnvironmentSecretType
		name  string
	}

	// Two keys of one group under the same type and name: the later key wins, as the merge in the task
	// runner let the last row win. Keeping both is impossible, the pair is unique in the new table.
	winners := map[slot]link{}
	var order []slot
	for _, k := range keys {
		var typ db.EnvironmentSecretType
		switch db.AccessKeyOwner(k.Owner) {
		case db.AccessKeyVariable:
			typ = db.EnvironmentSecretVar
		case db.AccessKeyEnvironment:
			typ = db.EnvironmentSecretEnv
		default:
			continue
		}
		name := strings.TrimPrefix(k.Name, string(typ)+".")
		s := slot{envID: k.EnvironmentID, typ: typ, name: name}
		if prev, dup := winners[s]; dup {
			log.WithFields(log.Fields{
				"context":        "migration",
				"version":        "2.20.18",
				"environment_id": k.EnvironmentID,
				"type":           typ,
				"name":           name,
				"kept_key_id":    k.ID,
				"dropped_key_id": prev.keyID,
			}).Warn("A variable group holds two secrets with the same name; only the newer key is linked")
		} else {
			order = append(order, s)
		}
		winners[s] = link{keyID: k.ID, projectID: k.ProjectID, envID: k.EnvironmentID, secretTyp: typ, name: name}
	}

	for _, s := range order {
		l := winners[s]
		_, err = tx.Exec(m.db.PrepareQuery(
			"insert into project__environment_secret (project_id, environment_id, type, name, key_id) values (?, ?, ?, ?, ?)"),
			l.projectID, l.envID, string(l.secretTyp), l.name, l.keyID)
		if err != nil {
			return err
		}
	}

	return nil
}
