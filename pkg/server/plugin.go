package server

import (
	"fmt"
	"os"
)

func PluginDir() string {
	home := os.Getenv("HOME")
	dir := fmt.Sprintf("%s/.kaytu/plugins/", home)
	os.MkdirAll(dir, os.ModePerm)
	return dir
}

func LogsDir() string {
	home := os.Getenv("HOME")
	dir := fmt.Sprintf("%s/.kaytu/logs", home)
	os.MkdirAll(dir, os.ModePerm)
	return dir
}
