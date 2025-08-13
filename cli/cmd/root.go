package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/services/server"

	"github.com/gorilla/handlers"
	"github.com/semaphoreui/semaphore/api"
	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/factory"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	proServer "github.com/semaphoreui/semaphore/pro/services/server"
	proTasks "github.com/semaphoreui/semaphore/pro/services/tasks"
	"github.com/semaphoreui/semaphore/services/schedules"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
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
		str := persistentFlags.logLevel
		if str == "" {
			str = os.Getenv("SEMAPHORE_LOG_LEVEL")
		}
		if str == "" {
			return
		}

		lvl, err := log.ParseLevel(str)
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("Log level set to", lvl)
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
	state := proTasks.NewTaskStateStore()
	terraformStore := proFactory.NewTerraformStore(store)
	ansibleTaskRepo := proFactory.NewAnsibleTaskRepository(store)

	projectService := server.NewProjectService(store, store)
	encryptionService := server.NewAccessKeyEncryptionService(store, store, store)
	accessKeyInstallationService := server.NewAccessKeyInstallationService(encryptionService)
	integrationService := server.NewIntegrationService(store, encryptionService)
	inventoryService := server.NewInventoryService(
		store,
		store,
		store,
		encryptionService,
	)
	accessKeyService := server.NewAccessKeyService(store, encryptionService, store)
	secretStorageService := server.NewSecretStorageService(store, accessKeyService)
	environmentService := server.NewEnvironmentService(store, encryptionService)
	subscriptionService := proServer.NewSubscriptionService(store, store)
	logWriteService := proServer.NewLogWriteService()

	taskPool := tasks.CreateTaskPool(
		store,
		state,
		ansibleTaskRepo,
		inventoryService,
		encryptionService,
		accessKeyInstallationService,
		logWriteService,
	)

	schedulePool := schedules.CreateSchedulePool(
		store,
		&taskPool,
		accessKeyInstallationService,
		encryptionService,
	)

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

	subscriptionService.StartValidationCron()

	go sockets.StartWS()
	go schedulePool.Run()
	go taskPool.Run()

	route := api.Route(
		store,
		terraformStore,
		ansibleTaskRepo,
		&taskPool,
		projectService,
		integrationService,
		encryptionService,
		accessKeyInstallationService,
		secretStorageService,
		accessKeyService,
		environmentService,
		subscriptionService,
	)

	route.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = helpers.SetContextValue(r, "store", store)
			r = helpers.SetContextValue(r, "schedule_pool", schedulePool)
			r = helpers.SetContextValue(r, "task_pool", &taskPool)
			r = helpers.SetContextValue(r, "log_writer", logWriteService)
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

	var err error
	if util.Config.TLS.Enabled {
		if util.Config.TLS.HTTPRedirectPort != nil {

			go func() {
				httpRedirectPort := fmt.Sprintf(":%d", *util.Config.TLS.HTTPRedirectPort)
				err = http.ListenAndServe(httpRedirectPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					target := "https://"

					if util.Config.WebHost != "" {
						webHost, err2 := url.Parse(util.Config.WebHost)
						if err2 != nil {
							log.Panic(err2)
						}
						target += webHost.Host + r.URL.Path
					} else {
						hostParts := strings.Split(r.Host, ":")
						host := hostParts[0]
						target += host + port + r.URL.Path
					}

					if len(r.URL.RawQuery) > 0 {
						target += "?" + r.URL.RawQuery
					}

					if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
						http.Error(w, "http requests forbidden", http.StatusForbidden)
						return
					}

					http.Redirect(w, r, target, http.StatusTemporaryRedirect)
				}))
				if err != nil {
					log.Panic(err)
				}
			}()
		}

		err = http.ListenAndServeTLS(util.Config.Interface+port, util.Config.TLS.CertFile, util.Config.TLS.KeyFile, cropTrailingSlashMiddleware(router))

		if err != nil {
			log.Panic(err)
		}

	} else {
		err = http.ListenAndServe(util.Config.Interface+port, cropTrailingSlashMiddleware(router))
	}

	if err != nil {
		log.WithError(err).Panic("Error starting server")
	}
}

func createStoreWithMigrationVersion(token string, undoTo *string, applyTo *string) db.Store {
	util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	store := factory.CreateStore()

	store.Connect(token)

	var err error
	if undoTo != nil {
		err = db.Rollback(store, *undoTo)
	} else {
		err = db.Migrate(store, applyTo)
	}

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

func createStore(token string) db.Store {
	return createStoreWithMigrationVersion(token, nil, nil)
}
