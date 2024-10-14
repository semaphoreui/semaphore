package db

import (
	"fmt"
	"time"
)

// Migration represents sql schema version
type Migration struct {
	Version      string     `db:"version" json:"version"`
	UpgradedDate *time.Time `db:"upgraded_date" json:"upgraded_date"`
	Notes        *string    `db:"notes" json:"notes"`
}

// HumanoidVersion adds a v to the VersionString
func (m Migration) HumanoidVersion() string {
	return "v" + m.Version
}

func GetMigrations() []Migration {
	return []Migration{
		{Version: "0.0.0"},
		{Version: "1.0.0"},
		{Version: "1.2.0"},
		{Version: "1.3.0"},
		{Version: "1.4.0"},
		{Version: "1.5.0"},
		{Version: "1.6.0"},
		{Version: "1.7.0"},
		{Version: "1.8.0"},
		{Version: "1.9.0"},
		{Version: "2.2.1"},
		{Version: "2.3.0"},
		{Version: "2.3.1"},
		{Version: "2.3.2"},
		{Version: "2.4.0"},
		{Version: "2.5.0"},
		{Version: "2.5.2"},
		{Version: "2.7.1"},
		{Version: "2.7.4"},
		{Version: "2.7.6"},
		{Version: "2.7.8"},
		{Version: "2.7.9"},
		{Version: "2.7.10"},
		{Version: "2.7.12"},
		{Version: "2.7.13"},
		{Version: "2.8.0"},
		{Version: "2.8.1"},
		{Version: "2.8.7"},
		{Version: "2.8.8"},
		{Version: "2.8.20"},
		{Version: "2.8.25"},
		{Version: "2.8.26"},
		{Version: "2.8.36"},
		{Version: "2.8.38"},
		{Version: "2.8.39"},
		{Version: "2.8.40"},
		{Version: "2.8.42"},
		{Version: "2.8.51"},
		{Version: "2.8.57"},
		{Version: "2.8.58"},
		{Version: "2.8.91"},
		{Version: "2.9.6"},
		{Version: "2.9.46"},
		{Version: "2.9.60"},
		{Version: "2.9.61"},
		{Version: "2.9.62"},
		{Version: "2.9.70"},
		{Version: "2.9.97"},
		{Version: "2.9.100"},
		{Version: "2.10.12"},
		{Version: "2.10.15"},
		{Version: "2.10.16"},
		{Version: "2.10.24"},
		{Version: "2.10.26"},
	}
}

func Migrate(d Store) error {
	didRun := false

	for _, version := range GetMigrations() {
		if exists, err := d.IsMigrationApplied(version); err != nil || exists {
			if exists {
				continue
			}

			return err
		}

		didRun = true
		fmt.Printf("Executing migration %s (at %v)...\n", version.HumanoidVersion(), time.Now())
		if err := d.ApplyMigration(version); err != nil {
			fmt.Printf("Rolling back %s (time: %v)...\n", version.HumanoidVersion(), time.Now())
			d.TryRollbackMigration(version)
			return err
		}
	}

	if didRun {
		fmt.Println("Migrations Finished")
	}

	return nil
}
