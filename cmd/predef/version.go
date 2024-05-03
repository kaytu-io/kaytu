package predef

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VERSION string

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(GetVersion())
	},
}

func GetVersion() string {
	return VERSION
}
