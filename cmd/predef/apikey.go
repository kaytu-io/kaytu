package predef

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/api/kaytu"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use: "generate",
}

var ApiKeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "Generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := server.GetConfig()
		if err != nil {
			return err
		}

		resp, err := kaytu.ApiKeyRequest(cfg.AccessToken)
		if err != nil {
			return err
		}

		fmt.Println("New API Key generated:")
		fmt.Println(resp.Token)

		return nil
	},
}
