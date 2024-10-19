package cmd

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	for _, cmd := range []*cobra.Command{userChangeByLoginCmd, userChangeByEmailCmd} {
		cmd.PersistentFlags().StringVar(&targetUserArgs.login, "login", "", "User login")
		cmd.PersistentFlags().StringVar(&targetUserArgs.name, "name", "", "User's new name")
		cmd.PersistentFlags().StringVar(&targetUserArgs.email, "email", "", "User's new email")
		cmd.PersistentFlags().StringVar(&targetUserArgs.password, "password", "", "User's new password")
		cmd.PersistentFlags().BoolVar(&targetUserArgs.admin, "admin", false, "Mark user as admin")
		userCmd.AddCommand(cmd)
	}
}

func applyChangeUserArgsForUser(user db.User, store db.Store) {
	if targetUserArgs.name != "" {
		user.Name = targetUserArgs.name
	}

	if targetUserArgs.email != "" {
		user.Email = targetUserArgs.email
	}

	if targetUserArgs.login != "" {
		user.Username = targetUserArgs.login
	}

	if targetUserArgs.name != "" {
		user.Name = targetUserArgs.name
	}

	if targetUserArgs.admin {
		user.Admin = true
	}

	if err := store.UpdateUser(db.UserWithPwd{
		User: user,
		Pwd:  targetUserArgs.password,
	}); err != nil {
		panic(err)
	}

	fmt.Printf("User %s <%s> changed!\n", user.Username, user.Email)
}

var userChangeByLoginCmd = &cobra.Command{
	Use:   "change-by-login",
	Short: "Change user found by login",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.login == "" {
			fmt.Println("Argument --login required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user change-by-login --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, "")

		if err != nil {
			panic(err)
		}

		applyChangeUserArgsForUser(user, store)
	},
}

var userChangeByEmailCmd = &cobra.Command{
	Use:   "change-by-email",
	Short: "Change user found by email",
	Run: func(cmd *cobra.Command, args []string) {

		ok := true

		if targetUserArgs.email == "" {
			fmt.Println("Argument --email required")
			ok = false
		}

		if !ok {
			fmt.Println("Use command `semaphore user change-by-email --help` for details.")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail("", targetUserArgs.email)
		if err != nil {
			panic(err)
		}

		applyChangeUserArgsForUser(user, store)
	},
}
