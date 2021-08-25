package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/factory"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
	"os"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "semaphore",
	Short: "Ansible Semaphore is a beautiful web UI for Ansible",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at https://ansible-semaphore.com`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && configPath == "" {
			_ = cmd.Help()
			os.Exit(0)
		} else {
			serviceCmd.Run(cmd, args)
		}
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createStore() db.Store {
	util.ConfigInit(configPath)

	switch {
	case util.Config.MySQL.IsPresent():
		fmt.Printf("MySQL %v@%v %v\n", util.Config.MySQL.Username, util.Config.MySQL.Hostname, util.Config.MySQL.DbName)
	case util.Config.BoltDb.IsPresent():
		fmt.Printf("BoltDB %v\n", util.Config.BoltDb.Hostname)
	case util.Config.Postgres.IsPresent():
		fmt.Printf("Postgres %v@%v %v\n", util.Config.Postgres.Username, util.Config.Postgres.Hostname, util.Config.Postgres.DbName)
	default:
		panic(fmt.Errorf("database configuration not found"))
	}

	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)

	store := factory.CreateStore()

	if err := store.Connect(); err != nil {
		fmt.Println("\n Have you run `semaphore setup`?")
		panic(err)
	}

	if err := store.Migrate(); err != nil {
		panic(err)
	}

	return store
}