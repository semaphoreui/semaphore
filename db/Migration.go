package db

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/util"
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

func GetMigrations(dialect string) []Migration {

	var initScripts []Migration
	if dialect == util.DbDriverSQLite {
		initScripts = []Migration{{Version: "2.15.1.sqlite"}}
	} else {
		initScripts = []Migration{
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
			{Version: "2.10.28"},
			{Version: "2.10.33"},
			{Version: "2.10.46"},
			{Version: "2.11.5"},
			{Version: "2.12.0"},
			{Version: "2.12.3"},
			{Version: "2.12.4"},
			{Version: "2.12.5"},
			{Version: "2.12.15"},
			{Version: "2.13.0"},
			{Version: "2.14.0"},
			{Version: "2.14.1"},
			{Version: "2.14.5"},
			{Version: "2.14.7"},
			{Version: "2.14.12"},
			{Version: "2.15.0"},
			{Version: "2.15.1"},
		}
	}

	// add any new scripts here
	commonScripts := []Migration{
		{Version: "2.15.2"},
		{Version: "2.15.3"},
		{Version: "2.15.4"},
		{Version: "2.16.0"},
		{Version: "2.16.1"},
		{Version: "2.16.2"},
		{Version: "2.16.3"},
		{Version: "2.16.8"},
		{Version: "2.17.0"},
		{Version: "2.17.1"},
		{Version: "2.17.2"},
	}

	return append(initScripts, commonScripts...)
}

func (m Migration) Validate() error {
	if m.Version == "" {
		return fmt.Errorf("migration version is empty")
	}

	return nil
}

type MigrationVersion struct {
	Major int
	Minor int
	Patch int
}

func (m Migration) ParseVersion() (res MigrationVersion, err error) {

	parts := strings.Split(m.Version, ".")

	if len(parts) < 2 {
		err = fmt.Errorf("invalid migration version format %s", m.Version)
		return
	}

	res.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("invalid migration version major part %s", parts[0])
		return
	}

	res.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("invalid migration version minor part %s", parts[1])
		return
	}

	if len(parts) < 3 {
		return
	}

	res.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("invalid migration version patch part %s", parts[2])
		return
	}

	return
}

func (v MigrationVersion) Compare(o MigrationVersion) int {
	if v.Major < o.Major {
		return -1
	} else if v.Major > o.Major {
		return 1
	}

	if v.Minor < o.Minor {
		return -1
	} else if v.Minor > o.Minor {
		return 1
	}

	if v.Patch < o.Patch {
		return -1
	} else if v.Patch > o.Patch {
		return 1
	}

	return 0
}

func (m Migration) Compare(o Migration) int {

	mVer, err := m.ParseVersion()
	if err != nil {
		panic(err)
	}

	oVer, err := o.ParseVersion()
	if err != nil {
		panic(err)
	}

	return mVer.Compare(oVer)
}

func Rollback(d Store, targetVersion string) error {

	didRun := false

	migrations := GetMigrations(d.GetDialect())
	slices.Reverse(migrations)

	for _, version := range migrations {

		if version.Compare(Migration{Version: targetVersion}) <= 0 {
			break
		}

		applied, err := d.IsMigrationApplied(version)
		if err != nil {
			return err
		}

		if !applied {
			continue
		}

		d.TryRollbackMigration(version)

		didRun = true
	}

	if didRun {
		fmt.Println("Rollback Finished")
	}

	return nil
}

func Migrate(d Store, targetVersion *string) error {
	didRun := false

	for _, version := range GetMigrations(d.GetDialect()) {

		if targetVersion != nil && version.Compare(Migration{Version: *targetVersion}) > 0 {
			break
		}

		if exists, err := d.IsMigrationApplied(version); err != nil || exists {
			if exists {
				continue
			}

			return err
		}

		didRun = true
		fmt.Printf("Executing migration %s (at %v)...\n", version.HumanoidVersion(), tz.Now())
		if err := d.ApplyMigration(version); err != nil {
			fmt.Printf("Rolling back %s (time: %v)...\n", version.HumanoidVersion(), tz.Now())
			d.TryRollbackMigration(version)
			return err
		}
	}

	if didRun {
		fmt.Println("Migrations Finished")
	}

	return nil
}
