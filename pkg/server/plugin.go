package server

import (
	"encoding/json"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Plugin struct {
	Config *golang.RegisterConfig `json:"config"`
}

func (p *Plugin) Path() string {
	executableName := strings.ReplaceAll(p.Config.Name, "/", "_")
	_ = filepath.WalkDir(PluginDir(), func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) == executableName {
			executableName = info.Name()
			return nil
		}
		return nil
	})

	return filepath.Join(PluginDir(), executableName)
}

func GetPlugins() ([]*Plugin, error) {
	var cfg Config
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("[GetPlugins] : %v", err)
	}
	path := filepath.Join(home, ".kaytu", "kaytu-config.json")

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// if the file does not exist, return nil
			return nil, nil
		}
		return nil, fmt.Errorf("[GetPlugins] : %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("[GetPlugins] : %v", err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("[GetPlugins] : %v", err)
	}

	return cfg.Plugins, nil
}

func PluginDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".kaytu", "plugins")
	os.MkdirAll(dir, os.ModePerm)
	return dir
}

func LogsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".kaytu", "logs")
	os.MkdirAll(dir, os.ModePerm)
	return dir
}
