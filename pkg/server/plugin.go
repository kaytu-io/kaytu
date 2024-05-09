package server

import (
	"os"
	"path"
)

func PluginDir() string {
	home, _ := os.UserHomeDir()
	dir := path.Join(home, ".kaytu", "plugins")
	os.MkdirAll(dir, os.ModePerm)
	return dir
}

func LogsDir() string {
	home, _ := os.UserHomeDir()
	dir := path.Join(home, ".kaytu", "logs")
	os.MkdirAll(dir, os.ModePerm)
	return dir
}
