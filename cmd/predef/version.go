package predef

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/version"

	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Kaytu CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.VERSION)
	},
}
