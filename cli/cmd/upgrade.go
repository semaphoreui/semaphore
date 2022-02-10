package cmd

import (
	"errors"
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Adapted from github.com/apex/apex

// doUpgrade checks for an update, and if available downloads the binary and installs it
func doUpgrade() error {
	updateAvailable, err := util.CheckUpdate()

	if err != nil || updateAvailable == nil {
		return err
	}

	asset := findAsset(updateAvailable)
	if asset == nil {
		return errors.New("cannot find binary for your system")
	}

	// create tmp file
	tmpPath := filepath.Join(os.TempDir(), "semaphore-upgrade")
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755) //nolint: gas
	if err != nil {
		return err
	}

	// download binary
	fmt.Printf("downloading %s\n", *asset.BrowserDownloadURL)
	res, err := http.Get(*asset.BrowserDownloadURL)
	if err != nil {
		return err
	}

	defer res.Body.Close() //nolint: errcheck

	// copy it down
	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}

	// replace it
	cmdPath := util.FindSemaphore()
	if len(cmdPath) == 0 {
		return errors.New("cannot find semaphore binary")
	}

	fmt.Printf("replacing %s\n", cmdPath)
	err = os.Rename(tmpPath, cmdPath)
	if err != nil {
		return err
	}

	fmt.Println("visit https://github.com/ansible-semaphore/semaphore/releases for the changelog")
	go func() {
		time.Sleep(time.Second * 3)
		os.Exit(0)
	}()

	return nil
}

// findAsset returns the binary for this platform.
func findAsset(release *github.RepositoryRelease) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		suffix := fmt.Sprintf("_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		if strings.HasPrefix(*asset.Name, "semaphore_") &&
			strings.HasSuffix(*asset.Name, suffix) {
			return &asset
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to latest stable version",
	Run: func(cmd *cobra.Command, args []string) {
		err := doUpgrade()
		if err != nil {
			panic(err)
		}
	},
}
