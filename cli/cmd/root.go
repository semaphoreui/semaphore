package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/api"
	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/factory"
	"github.com/semaphoreui/semaphore/services/schedules"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var persistentFlags struct {
	configPath string
	noConfig   bool
	logLevel   string
}

var rootCmd = &cobra.Command{
	Use:   "semaphore",
	Short: "Semaphore UI is a beautiful web UI for Ansible",
	Long: `Semaphore UI is a beautiful web UI for Ansible.
Source code is available at https://github.com/semaphoreui/semaphore.
Complete documentation is available at https://semaphoreui.com.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if persistentFlags.logLevel == "" {
			return
		}

		lvl, err := log.ParseLevel(persistentFlags.logLevel)
		if err != nil {
			log.Panic(err)
		}

		log.SetLevel(lvl)
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(&persistentFlags.logLevel, "log-level", "", "Log level: DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
	rootCmd.PersistentFlags().StringVar(&persistentFlags.configPath, "config", "", "Configuration file path")
	rootCmd.PersistentFlags().BoolVar(&persistentFlags.noConfig, "no-config", false, "Don't use configuration file")
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
	util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

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
