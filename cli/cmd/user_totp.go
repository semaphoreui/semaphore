package cmd

import (
	"fmt"
	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	for _, cmd := range []*cobra.Command{totpEnableCmd, totpDisableCmd, totpShowCmd} {
		cmd.PersistentFlags().StringVar(&targetUserArgs.login, "login", "", "User login")
		totpCmd.AddCommand(cmd)
	}
	userCmd.AddCommand(totpCmd)
}

var totpCmd = &cobra.Command{
	Use:   "totp",
	Short: "Manage TOTP verification",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var totpEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable TOTP verification",
	Run: func(cmd *cobra.Command, args []string) {

		if targetUserArgs.login == "" {
			fmt.Println("Argument --login required")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, "")

		if err != nil {
			panic(err)
		}

		if user.Totp != nil {
			fmt.Println("TOTP already enabled")
			os.Exit(1)
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Semaphore",
			AccountName: user.Email,
		})

		if err != nil {
			panic(err)
		}

		totp, err := store.AddTotpVerification(user.ID, key.URL())
		if err != nil {
			panic(err)
		}

		fmt.Println()
		fmt.Println(totp.URL)
		fmt.Println()
		qrterminal.GenerateWithConfig(totp.URL, qrterminal.Config{
			Writer:    os.Stdout,
			Level:     qrterminal.L,
			BlackChar: qrterminal.BLACK,
			WhiteChar: qrterminal.WHITE,
			QuietZone: 2,
		})
		fmt.Println()
	},
}

var totpDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable TOTP verification",
	Run: func(cmd *cobra.Command, args []string) {

		if targetUserArgs.login == "" {
			fmt.Println("Argument --login required")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, "")

		if err != nil {
			panic(err)
		}

		if user.Totp == nil {
			fmt.Println("TOTP not enabled")
			os.Exit(1)
		}

		err = store.DeleteTotpVerification(user.ID, user.Totp.ID)
		if err != nil {
			panic(err)
		}
	},
}

var totpShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show TOTP details",
	Run: func(cmd *cobra.Command, args []string) {
		if targetUserArgs.login == "" {
			fmt.Println("Argument --login required")
			os.Exit(1)
		}

		store := createStore("")
		defer store.Close("")

		user, err := store.GetUserByLoginOrEmail(targetUserArgs.login, "")

		if err != nil {
			panic(err)
		}

		if user.Totp == nil {
			fmt.Println("TOTP disabled")
		} else {
			fmt.Println()
			fmt.Println(user.Totp.URL)
			fmt.Println()
			qrterminal.GenerateWithConfig(user.Totp.URL, qrterminal.Config{
				Writer:    os.Stdout,
				Level:     qrterminal.L,
				BlackChar: qrterminal.BLACK,
				WhiteChar: qrterminal.WHITE,
				QuietZone: 2,
			})
			fmt.Println()
		}
	},
}
