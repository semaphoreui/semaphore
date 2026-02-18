package sql

import (
	"strconv"
	"strings"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

type migration_2_18_0 struct {
	db *SqlDb
}

func (m migration_2_18_0) PostApply(tx *gorp.Transaction) error {
	results, err := tx.Select(
		db.Option{},
		m.db.PrepareQuery("select `key`, `value` from `option` where `key` like ?"),
		"apps.%",
	)
	if err != nil {
		return err
	}

	type appFields struct {
		fields map[string]string
	}

	apps := make(map[string]*appFields)

	for _, r := range results {
		opt := r.(*db.Option)
		parts := strings.SplitN(opt.Key, ".", 3)
		if len(parts) != 3 {
			continue
		}
		appID := parts[1]
		field := parts[2]

		if apps[appID] == nil {
			apps[appID] = &appFields{fields: make(map[string]string)}
		}
		apps[appID].fields[field] = opt.Value
	}

	for appID, data := range apps {
		active := data.fields["active"] != "false"
		priority := 0
		if p, ok := data.fields["priority"]; ok {
			priority, _ = strconv.Atoi(p)
		}

		_, err = tx.Exec(
			m.db.PrepareQuery("insert into `app` (`id`, `title`, `icon`, `color`, `dark_color`, `active`, `priority`) values (?, ?, ?, ?, ?, ?, ?)"),
			appID,
			data.fields["title"],
			data.fields["icon"],
			data.fields["color"],
			data.fields["dark_color"],
			active,
			priority,
		)
		if err != nil {
			return err
		}

		args := data.fields["args"]

		_, err = tx.Exec(
			m.db.PrepareQuery("insert into `app__version` (`app_id`, `name`, `path`, `args`, `active`, `priority`) values (?, ?, ?, ?, ?, ?)"),
			appID,
			"",
			data.fields["path"],
			args,
			true,
			0,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(m.db.PrepareQuery("delete from `option` where `key` like ?"), "apps.%")
	return err
}
