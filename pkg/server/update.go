package server

import (
	"context"
	"fmt"
	githubAPI "github.com/google/go-github/v62/github"
	"github.com/kaytu-io/kaytu/pkg/version"
	"github.com/rogpeppe/go-internal/semver"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func CheckForUpdate() error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	if cfg.LastVersion != "" && semver.Compare(cfg.LastVersion, "v"+version.VERSION) > 0 {
		os.Stderr.WriteString("There's a new version for Kaytu CLI. Update it to latest version and enjoy the new features.\n")
		time.Sleep(2 * time.Second)
		return nil
	}

	if cfg.LastUpdateCheck.After(time.Now().Add(-7 * 24 * time.Hour)) {
		return nil
	}

	api := githubAPI.NewClient(nil)
	release, _, err := api.Repositories.GetLatestRelease(context.Background(), "kaytu-io", "kaytu")
	if err != nil {
		return err
	}

	for _, asset := range release.Assets {
		pattern := fmt.Sprintf("kaytu_([a-z0-9\\.]+)_%s_%s", runtime.GOOS, runtime.GOARCH)
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if asset.Name != nil && r.MatchString(*asset.Name) {
			ver := strings.Split(*asset.Name, "_")[1]
			cfg.LastVersion = ver
			cfg.LastUpdateCheck = time.Now()
			err = SetConfig(*cfg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
