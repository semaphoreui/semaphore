package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
	"os"
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

		ok := true
		if newUserArgs.name == "" {
			fmt.Println("Argument --name requied")
			ok = false
		}
		if newUserArgs.login == "" {
			fmt.Println("Argument --login requied")
			ok = false
		}

		if newUserArgs.email == "" {
			fmt.Println("Argument --email requied")
			ok = false
		}

		if newUserArgs.password == "" {
			fmt.Println("Argument --password requied")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user add --help` for details.")
			os.Exit(1)
		}

		store := createStore()
		defer store.Close()

		if _, err := store.CreateUser(db.UserWithPwd{
			Pwd: newUserArgs.password,
			User: db.User{
				Name: newUserArgs.name,
				Username: newUserArgs.login,
				Email: newUserArgs.email,
				Admin: true,
			},
		}); err != nil {
			panic(err)
		}

		fmt.Printf("User %s <%s> added!", newUserArgs.login, newUserArgs.email)
	},
}