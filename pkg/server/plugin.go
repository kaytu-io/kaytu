package server

import (
	"os"
	"path/filepath"
)

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
