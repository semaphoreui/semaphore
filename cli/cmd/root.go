package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/api"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/factory"
	"github.com/ansible-semaphore/semaphore/services/schedules"
	"github.com/ansible-semaphore/semaphore/services/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strings"
)

var configPath string
var noConfig bool

var rootCmd = &cobra.Command{
	Use:   "semaphore",
	Short: "Semaphore UI is a beautiful web UI for Ansible",
	Long: `Semaphore UI is a beautiful web UI for Ansible.
Source code is available at https://github.com/ansible-semaphore/semaphore.
Complete documentation is available at https://ansible-semaphore.com.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file path")
	rootCmd.PersistentFlags().BoolVar(&noConfig, "no-config", false, "Don't use configuration file")
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runService() {
	store := createStore("root")
	taskPool := tasks.CreateTaskPool(store)
	schedulePool := schedules.CreateSchedulePool(store, &taskPool)

	defer schedulePool.Destroy()

	util.Config.PrintDbInfo()

	port := util.Config.Port

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)
	fmt.Printf("Semaphore %v\n", util.Version())
	fmt.Printf("Interface %v\n", util.Config.Interface)
	fmt.Printf("Port %v\n", util.Config.Port)

	go sockets.StartWS()
	go schedulePool.Run()
	go taskPool.Run()

	route := api.Route()

	route.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "store", store)
			context.Set(r, "schedule_pool", schedulePool)
			context.Set(r, "task_pool", &taskPool)
			next.ServeHTTP(w, r)
		})
	})

	var router http.Handler = route

	router = handlers.ProxyHeaders(router)
	http.Handle("/", router)

	fmt.Println("Server is running")

	if store.PermanentConnection() {
		defer store.Close("root")
	} else {
		store.Close("root")
	}

	err := http.ListenAndServe(util.Config.Interface+port, cropTrailingSlashMiddleware(router))

	if err != nil {
		log.Panic(err)
	}
}

func createStore(token string) db.Store {
	util.ConfigInit(configPath, noConfig)

	store := factory.CreateStore()

	store.Connect(token)

	err := db.Migrate(store)

	if err != nil {
		panic(err)
	}

	err = db.FillConfigFromDB(store)

	if err != nil {
		panic(err)
	}

	util.LookupDefaultApps()

	return store
}
