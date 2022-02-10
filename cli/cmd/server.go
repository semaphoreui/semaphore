package cmd

import (
	"github.com/spf13/cobra"
	"net/http"
	"strings"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Run in server mode",
	Aliases: []string{"service"},
	Run: func(cmd *cobra.Command, args []string) {
		runService()
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
