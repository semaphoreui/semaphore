package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/factory"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "semaphore",
	Short: "Ansible Semaphore is a beautiful web UI for Ansible",
	Long: 	`Ansible Semaphore is a beautiful web UI for Ansible.
Source code is available at https://github.com/ansible-semaphore/semaphore.
Complete documentation is available at https://ansible-semaphore.com.`,
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
	args := os.Args[1:]
	if len(args) == 2 && args[0] == "-config" {
		configPath = args[1]
		runService()
		return
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runService() {
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

	store := createStore()
	defer store.Close()

	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Interface %v\n", util.Config.Interface)
	fmt.Printf("Port %v\n", util.Config.Port)

	go sockets.StartWS()
	go tasks.StartRunner()

	route := api.Route()

	route.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "store", store)
			next.ServeHTTP(w, r)
		})
	})

	var router http.Handler = route

	router = handlers.ProxyHeaders(router)
	http.Handle("/", router)

	fmt.Println("Server is running")

	err := http.ListenAndServe(util.Config.Interface+util.Config.Port, cropTrailingSlashMiddleware(router))

	if err != nil {
		log.Panic(err)
	}
}

func createStore() db.Store {
	util.ConfigInit(configPath)

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