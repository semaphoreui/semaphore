package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(userListCmd)
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all users",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore("")
		defer store.Close("")

		users, err := store.GetUsers(db.RetrieveQueryParams{})

		if err != nil {
			panic(err)
		}

		for _, user := range users {
			fmt.Println(user.Username)
		}
	},
}
