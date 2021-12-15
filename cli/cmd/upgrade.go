package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Adapted from github.com/apex/apex

// doUpgrade checks for an update, and if available downloads the binary and installs it
func doUpgrade(currentVersion string) error {
	fmt.Printf("current release is v%s\n", currentVersion)

	updateAvailable, err := checkUpdate(currentVersion)

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
	cmdPath := findSemaphore()
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

// findSemaphore looks in the PATH for the semaphore variable
// if not found it will attempt to find the absolute path of the first
// os argument, the semaphore command, and return it
func findSemaphore() string {
	cmdPath, _ := exec.LookPath("semaphore") //nolint: gas

	if len(cmdPath) == 0 {
		cmdPath, _ = filepath.Abs(os.Args[0]) // nolint: gas
	}

	return cmdPath
}

// findAsset returns the binary for this platform.
func findAsset(release *github.RepositoryRelease) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if *asset.Name == fmt.Sprintf("semaphore_%s_%s", runtime.GOOS, runtime.GOARCH) {
			return asset
		}
	}

	return nil
}

// checkUpdate uses the GitHub client to check for new tags in the semaphore repo
func checkUpdate(currentVersion string) (updateAvailable *github.RepositoryRelease, err error) {
	// fetch releases
	gh := github.NewClient(nil)
	releases, _, err := gh.Repositories.ListReleases(context.TODO(), "ansible-semaphore", "semaphore", nil)
	if err != nil {
		return
	}

	updateAvailable = nil
	if (*releases[0].TagName)[1:] != currentVersion {
		updateAvailable = releases[0]
	}

	return
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to latest stable version",
	Run: func(cmd *cobra.Command, args []string) {
		err := doUpgrade(util.Version)
		if err != nil {
			panic(err)
		}
	},
}
