package flags

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func ReadStringFlag(cmd *cobra.Command, name string) string {
	name = strings.ReplaceAll(strcase.ToSnake(name), "_", "-")
	if cmd.Flags().Lookup(name) == nil {
		fmt.Println("cant find", name)
	}
	value := cmd.Flags().Lookup(name).Value.String()
	if strings.HasPrefix(value, "@") {
		return readFile(value[1:])
	} else if strings.HasPrefix(value, "file://") {
		return readFile(value[7:])
	}
	return value
}

func readFile(path string) string {
	var fullPath string

	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fullPath = filepath.Join(wd, path)
	}

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	return string(content)
}
