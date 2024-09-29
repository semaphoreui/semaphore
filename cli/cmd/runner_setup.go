package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/cli/setup"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/spf13/cobra"
)

func init() {
	runnerCmd.AddCommand(runnerSetupCmd)
}

var runnerSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Perform interactive setup",
	Run: func(cmd *cobra.Command, args []string) {
		doRunnerSetup()
	},
}

// nolint: gocyclo
func doRunnerSetup() int {
	var config *util.ConfigType
	config = &util.ConfigType{}

	setup.InteractiveRunnerSetup(config)

	configPath := setup.SaveConfig(config)

	util.ConfigInit(configPath, false)

	fmt.Printf(" Re-launch this program pointing to the configuration file\n\n./semaphore server --config %v\n\n", configPath)
	fmt.Printf(" To run as daemon:\n\nnohup ./semaphore server --config %v &\n\n", configPath)

	return 0
}
