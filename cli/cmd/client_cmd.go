package cmd

import "github.com/semaphoreui/semaphore/cli/cmd/client"

// The REST API client lives in its own package so that it never reaches the
// database or util.Config; it is attached here so that RootCommand() and the
// docs generator see it.
func init() {
	rootCmd.AddCommand(client.NewCommand(client.StdIO()))
}
