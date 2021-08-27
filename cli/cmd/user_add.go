package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
)

var newUserArgs userArgs

func init() {
	userAddCmd.PersistentFlags().StringVar(&newUserArgs.login, "login", "", "New user login")
	userAddCmd.PersistentFlags().StringVar(&newUserArgs.name, "name", "", "New user name")
	userAddCmd.PersistentFlags().StringVar(&newUserArgs.email, "email", "", "New user email")
	userAddCmd.PersistentFlags().StringVar(&newUserArgs.password, "password", "", "New user password")
	userCmd.AddCommand(userAddCmd)
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new user",
	Run: func(cmd *cobra.Command, args []string) {
		store := createStore()
		defer store.Close()

		if _, err := store.CreateUser(db.UserWithPwd{
			Pwd: newUserArgs.password,
			User: db.User{
				Name: newUserArgs.name,
				Username: newUserArgs.login,
				Email: newUserArgs.email,
			},
		}); err != nil {
			panic(err)
		}

		fmt.Printf("User %s <%s> added!", newUserArgs.login, newUserArgs.email)
	},
}