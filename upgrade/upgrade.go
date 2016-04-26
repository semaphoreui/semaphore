package upgrade

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/github"
)

// Adapted from github.com/apex/apex

func Upgrade(version string) error {
	fmt.Printf("current release is v%s\n", version)

	// fetch releases
	gh := github.NewClient(nil)
	releases, _, err := gh.Repositories.ListReleases("ansible-semaphore", "semaphore", nil)
	if err != nil {
		return err
	}

	// see if it's new
	latest := releases[0]
	fmt.Printf("latest release is %s\n", *latest.TagName)

	if (*latest.TagName)[1:] == version {
		return nil
	}

	asset := findAsset(&latest)
	if asset == nil {
		return errors.New("cannot find binary for your system")
	}

	// create tmp file
	tmpPath := filepath.Join(os.TempDir(), "semaphore-upgrade")
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	// download binary
	fmt.Printf("downloading %s\n", *asset.BrowserDownloadURL)
	res, err := http.Get(*asset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// copy it down
	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}

	// replace it
	cmdPath, err := exec.LookPath("semaphore")
	if err != nil {
		return err
	}

	fmt.Printf("replacing %s\n", cmdPath)
	err = os.Rename(tmpPath, cmdPath)
	if err != nil {
		return err
	}

	fmt.Println("visit https://github.com/ansible-semaphore/semaphore/releases for the changelog")
	return nil
}

// findAsset returns the binary for this platform.
func findAsset(release *github.RepositoryRelease) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if *asset.Name == fmt.Sprintf("semaphore_%s_%s", runtime.GOOS, runtime.GOARCH) {
			return &asset
		}
	}

	return nil
}
