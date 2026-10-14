package sql

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/go-gorp/gorp/v3"
	log "github.com/sirupsen/logrus"
)

type migration_2_20_12 struct {
	db *SqlDb
}

// PostApply logs the grants the rebuild left out, drops the old tables and renames slugs that collide by scope.
func (m migration_2_20_12) PostApply(tx *gorp.Transaction) error {
	var grants []struct {
		ID       int    `db:"id"`
		RoleSlug string `db:"role_slug"`
	}
	_, err := tx.Select(&grants, m.db.PrepareQuery(`
		select
		  tr.id, tr.role_slug
		from project__template_role_old tr
		where not exists (select 1 from project__template_role t where t.id = tr.id)
	`))
	if err != nil {
		return err
	}
	for _, o := range grants {
		log.WithFields(log.Fields{"context": "migration", "version": "2.20.12", "template_role_id": o.ID, "role": o.RoleSlug}).
			Warn("Dropped a template permission whose role does not exist or belongs to another project")
	}

	for _, table := range []string{"project__template_role_old", "project__user_old", "role_old"} {
		if _, err = tx.Exec("drop table " + table); err != nil {
			return err
		}
	}

	var roles []struct {
		ID        int    `db:"id"`
		Slug      string `db:"slug"`
		ProjectID *int   `db:"project_id"`
		Builtin   bool   `db:"builtin"`
	}
	if _, err = tx.Select(&roles, m.db.PrepareQuery("select id, slug, project_id, builtin from `role`")); err != nil {
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

	taken := map[string]bool{}
	for _, r := range roles {
		taken[strings.ToLower(r.Slug)] = true
	}

	kept := map[string]bool{}
	for _, r := range roles {
		global := strings.ToLower(r.Slug)
		key := global
		if r.ProjectID != nil {
			key = fmt.Sprintf("%s/%d", global, *r.ProjectID)
		}
		if !kept[global] && !kept[key] {
			kept[key] = true
			continue
		}

		newSlug := freeRoleSlug(r.Slug, taken)
		if _, err = tx.Exec(m.db.PrepareQuery("update `role` set slug=? where id=?"), newSlug, r.ID); err != nil {
			return err
		}
		log.WithFields(log.Fields{"context": "migration", "version": "2.20.12", "old_slug": r.Slug, "new_slug": newSlug}).
			Warn("Renamed a role whose slug collides with another role ignoring case or with a built-in role")
	}
	return nil
}

func freeRoleSlug(slug string, taken map[string]bool) string {
	for n := 2; ; n++ {
		suffix := fmt.Sprintf("-%d", n)
		cut := min(len(slug), 100-len(suffix))
		for cut > 0 && cut < len(slug) && !utf8.RuneStart(slug[cut]) {
			cut--
		}
		candidate := slug[:cut] + suffix
		if !taken[strings.ToLower(candidate)] {
			taken[strings.ToLower(candidate)] = true
			return candidate
		}
	}
}
