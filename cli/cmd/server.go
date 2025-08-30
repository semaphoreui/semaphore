package cmd

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var serverFlags struct {
	mcpEnabled bool
	mcpPort    string
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().BoolVar(&serverFlags.mcpEnabled, "mcp-enabled", false, "Enable MCP server")
	serverCmd.PersistentFlags().StringVar(&serverFlags.mcpPort, "mcp-port", "", "Port for MCP server")
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
