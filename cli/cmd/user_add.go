package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
)

func init() {
	userCmd.AddCommand(userAddCmd)
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Print the version number of Semaphore",

	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore()
		defer store.Close()

		if _, err := store.CreateUser(db.UserWithPwd{
			Pwd: UserAdd.Password,
			User: db.User{
				Name: UserAdd.Name,
				Username: UserAdd.Username,
				Email: UserAdd.Email,
			},
		}); err != nil {
			panic(err)
		}

		fmt.Printf("User %s <%s> added!", UserAdd.Username, UserAdd.Email)
	},
}