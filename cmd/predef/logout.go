package predef

import (
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Kaytu",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := server.RemoveConfig()
		if err != nil {
			return err
		}
		return nil
	},
}
