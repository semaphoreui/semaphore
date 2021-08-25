package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/spf13/cobra"
	"net/http"
	"strings"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run Ansible Semaphore service",
	// Long:  `All software has versions. This is Hugo's`,

	Run: func(cmd *cobra.Command, args []string) {
		store := createStore()
		defer store.Close()

		fmt.Printf("Semaphore %v\n", util.Version)
		fmt.Printf("Interface %v\n", util.Config.Interface)
		fmt.Printf("Port %v\n", util.Config.Port)

		go sockets.StartWS()
		//go checkUpdates()
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
	},
}

func cropTrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}

