package cmd

import (
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var migrationArgs struct {
	undoTo  string
	applyTo string
}

func init() {
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.undoTo, "undo-to", "", "Undo to specific version")
	migrateCmd.PersistentFlags().StringVar(&migrationArgs.undoTo, "apply-to", "", "Apply to specific version")

	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Execute migrations",
	Run: func(cmd *cobra.Command, args []string) {

		var undoTo, applyTo *string

		if migrationArgs.undoTo != "" && migrationArgs.applyTo != "" {
			panic("Cannot specify both --undo-to and --apply-to")
		}

		if migrationArgs.undoTo != "" {
			undoTo = &migrationArgs.undoTo
		}

		if migrationArgs.applyTo != "" {
			applyTo = &migrationArgs.applyTo
		}

		store := createStoreWithMigrationVersion("migrate", undoTo, applyTo)

		defer store.Close("migrate")
		util.Config.PrintDbInfo()
	},
}
