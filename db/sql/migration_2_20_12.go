package sql

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

type migration_2_20_12 struct {
	db *SqlDb
}

// PostApply finishes the scratch tables and swaps them in last, so on MySQL a failure leaves the old tables untouched.
func (m migration_2_20_12) PostApply(tx *gorp.Transaction) error {
	var grants []struct {
		ID       int    `db:"id"`
		RoleSlug string `db:"role_slug"`
	}
	_, err := tx.Select(&grants, m.db.PrepareQuery(`
		select
		  tr.id, tr.role_slug
		from project__template_role tr
		where not exists (select 1 from project__template_role_new t where t.id = tr.id)
	`))
	if err != nil {
		return err
	}
	for _, o := range grants {
		log.WithFields(log.Fields{"context": "migration", "version": "2.20.12", "template_role_id": o.ID, "role": o.RoleSlug}).
			Warn("Dropped a template permission whose role does not exist or belongs to another project")
	}

	// The database compares with built-in slugs itself: MySQL collations also ignore accents and trailing spaces.
	var roles []struct {
		ID             int    `db:"id"`
		Slug           string `db:"slug"`
		ProjectID      *int   `db:"project_id"`
		Builtin        bool   `db:"builtin"`
		ShadowsBuiltin bool   `db:"shadows_builtin"`
	}
	_, err = tx.Select(&roles, m.db.PrepareQuery(`
		select
		  id, slug, project_id, builtin,
		  builtin = false and slug in ('owner', 'manager', 'task_runner', 'guest') as shadows_builtin
		from role_new
	`))
	if err != nil {
		return err
	}

	// Built-in roles come first, then global roles, then the smallest slugs, so the earlier role keeps its slug.
	sort.Slice(roles, func(i, j int) bool {
		a, b := roles[i], roles[j]
		if a.Builtin != b.Builtin {
			return a.Builtin
		}
		if (a.ProjectID == nil) != (b.ProjectID == nil) {
			return a.ProjectID == nil
		}
		return a.Slug < b.Slug
	})

	kept := map[string]bool{}
	for _, r := range roles {
		global := strings.ToLower(r.Slug)
		key := global
		if r.ProjectID != nil {
			key = fmt.Sprintf("%s/%d", global, *r.ProjectID)
		}
		if !r.ShadowsBuiltin && !kept[global] && !kept[key] {
			kept[key] = true
			continue
		}

		// The database checks the candidates, so its collation decides what is taken, as in the unique index.
		var newSlug string
		for n := 2; newSlug == ""; n++ {
			suffix := fmt.Sprintf("-%d", n)
			runes := []rune(r.Slug)
			candidate := string(runes[:min(len(runes), 100-len(suffix))]) + suffix
			taken, err := tx.SelectInt(m.db.PrepareQuery("select count(*) from role_new where lower(slug) = lower(?)"), candidate)
			if err != nil {
				return err
			}
			if taken == 0 {
				newSlug = candidate
			}
		}
		if _, err = tx.Exec(m.db.PrepareQuery("update role_new set slug=? where id=?"), newSlug, r.ID); err != nil {
			return err
		}
		log.WithFields(log.Fields{"context": "migration", "version": "2.20.12", "old_slug": r.Slug, "new_slug": newSlug}).
			Warn("Renamed a role whose slug collides with another role ignoring case or with a built-in role")
	}

	swap := []string{
		"create unique index role__scope_slug on role_new (scope, lower(slug))",
		"alter table role rename to role_old",
		"alter table role_new rename to role",
		"alter table project__user rename to project__user_old",
		"alter table project__user_role_new rename to project__user_role",
		"alter table project__template_role rename to project__template_role_old",
		"alter table project__template_role_new rename to project__template_role",
	}
	if m.db.GetDialect() == util.DbDriverMySQL {
		// The ci collation already ignores case, and one rename statement swaps all tables at once.
		swap = []string{
			"create unique index role__scope_slug on role_new (scope, slug)",
			`
				rename table
				  role to role_old,
				  role_new to role,
				  project__user to project__user_old,
				  project__user_role_new to project__user_role,
				  project__template_role to project__template_role_old,
				  project__template_role_new to project__template_role
			`,
		}
	}
	for _, query := range swap {
		if _, err = tx.Exec(m.db.PrepareQuery(query)); err != nil {
			return err
		}
	}
	return nil
}
