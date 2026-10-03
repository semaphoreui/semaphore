package cmd

import "github.com/spf13/cobra"

// RootCommand exposes the assembled command tree so that tools outside this
// package can inspect it. tools/clidocs uses it to generate the CLI reference
// page of the documentation site, which is why the docs cannot describe a
// command that does not exist.
func RootCommand() *cobra.Command {
	registerPersistentFlags()
	return rootCmd
}
