package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/db/factory"
	"github.com/semaphoreui/semaphore/db/migration"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var migrationArgs struct {
	undoTo         string
	applyTo        string
	fromBoltDb     string
	errLogSize     int
	skipTaskOutput bool
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.undoTo, "undo-to", "", "Undo to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.applyTo, "apply-to", "", "Apply to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.fromBoltDb, "from-boltdb", "", "Path to boltDB data file")
	migrateCmd.PersistentFlags().IntVar(&migrationArgs.errLogSize, "err-log-size", 0, "Error log size")
	migrateCmd.PersistentFlags().BoolVar(&migrationArgs.skipTaskOutput, "skip-task-output", false, "Skip task output importing during migration")

	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute migrations",
	Run: func(cmd *cobra.Command, args []string) {

		if migrationArgs.undoTo != "" && migrationArgs.applyTo != "" {
			panic("Cannot specify both --undo-to and --apply-to")
		}

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

	if boltCfg.Dialect != util.DbDriverBolt {
		fmt.Printf("Error: Source database must be BoltDB (dialect: %s)\n", boltCfg.Dialect)
		return
	}

	file, err := os.Stat(boltDbPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("File does not exist")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	if file.Size() > 1024*1024*1024 {
		fmt.Println("File is too big ", file.Size())
	}

	boltStore := bolt.CreateBoltDB()
	boltStore.Filename = boltDbPath
	boltStore.Connect("migrate")

	defer boltStore.Close("migrate")

	util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	dialect, err := util.Config.GetDialect()
	if err != nil {
		fmt.Printf("Error reading SQL DB config: %v\n", err)
		return
	}

	if dialect == util.DbDriverBolt {
		fmt.Println("Error: Destination database must be a SQL database")
		return
	}

	sqlStore := factory.CreateStore()
	sqlStore.Connect("migrate")

	// 3. Connect and migrate
	fmt.Println("Starting migration...")
	migrator := &migration.Migrator{
		OldStore:       boltStore,
		NewStore:       sqlStore,
		ErrLogSize:     migrationArgs.errLogSize,
		SkipTaskOutput: migrationArgs.skipTaskOutput,
	}

	err = migrator.Migrate()
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		return
	}

	defer sqlStore.Close("migrate")

	fmt.Println("Migration finished successfully.")
}
