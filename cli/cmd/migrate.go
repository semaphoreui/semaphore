package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/db/factory"
	"github.com/semaphoreui/semaphore/db_migration"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var migrationArgs struct {
	undoTo     string
	applyTo    string
	fromBoltDb string
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.undoTo, "undo-to", "", "Undo to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.applyTo, "apply-to", "", "Apply to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.fromBoltDb, "from-boltdb", "", "Path to boltDB data file")
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute migrations",
	Run: func(cmd *cobra.Command, args []string) {

		if migrationArgs.undoTo != "" && migrationArgs.applyTo != "" {
			panic("Cannot specify both --undo-to and --apply-to")
		} else if migrationArgs.undoTo != "" || migrationArgs.applyTo != "" {
			var undoTo, applyTo *string

			if migrationArgs.undoTo != "" {
				undoTo = &migrationArgs.undoTo
			}

			if migrationArgs.applyTo != "" {
				applyTo = &migrationArgs.applyTo
			}

			store := createStoreWithMigrationVersion("migrate", undoTo, applyTo)

			defer store.Close("migrate")
			util.Config.PrintDbInfo()
		}

		if migrationArgs.fromBoltDb != "" {
			migrateBoltDb(migrationArgs.fromBoltDb)
		}
	},
}

func migrateBoltDb(boltDbPath string) {

	boltCfg := util.DbConfig{
		Dialect:  util.DbDriverBolt,
		Hostname: boltDbPath,
	}

	_, err := os.Stat(boltDbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("File does not exist")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	boltStore := bolt.CreateBoltDBWithConfig(boltCfg)
	boltStore.Connect("")

	// 2. Create SQL Store
	cfg, err := util.ConfigInitNew("", true, true)
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		return
	}
	sqlCfg, err := cfg.GetDBConfig()

	if err != nil {
		fmt.Printf("Error reading SQL DB config: %v\n", err)
		return
	}

	if sqlCfg.Dialect == util.DbDriverBolt {
		fmt.Println("Error: Destination database must be a SQL database")
		return
	}

	sqlStore := factory.CreateStoreWithConfig(sqlCfg)
	sqlStore.Connect("")

	// 3. Connect and migrate
	fmt.Println("Starting migration...")
	err = db_migration.Migrate(boltStore, sqlStore)
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		return
	}

	fmt.Println("Migration finished successfully.")
}
