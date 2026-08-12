package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/semaphoreui/semaphore/pkg/netproxy"
	"github.com/spf13/cobra"
)

var proxyConnectArgs struct {
	proxy string
}

func init() {
	proxyConnectCmd.PersistentFlags().StringVar(&proxyConnectArgs.proxy, "proxy", "",
		"proxy URL, for example socks5://user:pass@proxy.example.org:1080")

	rootCmd.AddCommand(proxyConnectCmd)
}

// proxyConnectCmd is used as the ssh ProxyCommand of a SOCKS5 or HTTP proxy. It
// is hidden because it is not something an administrator runs: it takes the
// place netcat would otherwise fill, without adding it to the image.
var proxyConnectCmd = &cobra.Command{
	Use:    "proxy-connect <host> <port>",
	Short:  "Connect stdio to a host through a SOCKS5 or HTTP proxy",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if proxyConnectArgs.proxy == "" {
			return fmt.Errorf("--proxy is required")
		}

		// Credentials come from the environment: a command line is readable by
		// every process on the host.
		creds := netproxy.Credentials{
			User:     os.Getenv("SEMAPHORE_PROXY_USER"),
			Password: os.Getenv("SEMAPHORE_PROXY_PASSWORD"),
		}

		conn, err := netproxy.Dial(context.Background(), proxyConnectArgs.proxy, args[0]+":"+args[1], creds)
		if err != nil {
			return err
		}

		defer conn.Close() //nolint: errcheck

		// ssh talks over stdio: whichever direction ends first ends the tunnel.
		done := make(chan error, 2)

		go func() { _, copyErr := io.Copy(conn, os.Stdin); done <- copyErr }()
		go func() { _, copyErr := io.Copy(os.Stdout, conn); done <- copyErr }()

		return <-done
	},
}
