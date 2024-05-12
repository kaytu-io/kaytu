package server

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/api/github"
	"github.com/kaytu-io/kaytu/pkg/version"
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

	if cfg.LastVersion != "" && cfg.LastVersion != version.VERSION {
		fmt.Println("There's a new version for Kaytu CLI. Update it to latest version and enjoy the new features.")
		time.Sleep(2 * time.Second)
		return nil
	}

	if cfg.LastUpdateCheck.After(time.Now().Add(-7 * 24 * time.Hour)) {
		return nil
	}

	release, err := github.GetLatestRelease("kaytu-io/kaytu")
	if err != nil {
		return err
	}

	for _, asset := range release.Assets {
		pattern := fmt.Sprintf("kaytu_([a-z0-9\\.]+)_%s_%s", runtime.GOOS, runtime.GOARCH)
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if r.MatchString(asset.Name) {
			ver := strings.Split(asset.Name, "_")[1]
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
