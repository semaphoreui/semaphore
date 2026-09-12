package cmd

import (
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var migrationArgs struct {
	undoTo             string
	applyTo            string
	errLogSize         int
	skipTaskOutput     bool
	mergeExistingUsers bool
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.undoTo, "undo-to", "", "Undo to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.applyTo, "apply-to", "", "Apply to specific version")
	migrateCmd.PersistentFlags().IntVar(&migrationArgs.errLogSize, "err-log-size", 0, "Error log size")
	migrateCmd.PersistentFlags().BoolVar(&migrationArgs.skipTaskOutput, "skip-task-output", false, "Skip task output importing during migration")
	migrateCmd.PersistentFlags().BoolVar(&migrationArgs.mergeExistingUsers, "merge-existing-users", false, "Reuse existing users matched by username instead of failing on conflict")

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

		defer store.Close()
		util.Config.PrintDbInfo()
	},
}
