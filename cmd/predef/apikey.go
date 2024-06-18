package predef

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/api/kaytu"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

var ApiKeyRootCmd = &cobra.Command{
	Use:   "apikey",
	Short: "Kaytu API key generate/list/delete",
}

var ApiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := server.GetConfig()
		if err != nil {
			return err
		}

		resp, err := kaytu.ListApiKeyRequest(cfg.AccessToken)
		if err != nil {
			return err
		}

		fmt.Println("List of all API keys:")
		fmt.Println("Name, CreatedAt, Active, Masked Key")
		for _, item := range resp {
			fmt.Printf("%s, %s, %s, %s\n", item.Name, item.CreatedAt.Format(time.RFC822), strconv.FormatBool(item.Active), item.MaskedKey)
		}

		return nil
	},
}

var ApiKeyCreateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generating Kaytu API key to use for Kaytu agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := server.GetConfig()
		if err != nil {
			return err
		}

		name := ""
		if len(args) == 0 || len(strings.TrimSpace(args[0])) == 0 {
			name = args[0]
		}

		resp, err := kaytu.CreateApiKeyRequest(cfg.AccessToken, name)
		if err != nil {
			return err
		}

		fmt.Println("New API Key generated:")
		fmt.Println(resp.Token)

		return nil
	},
}

var ApiKeyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := server.GetConfig()
		if err != nil {
			return err
		}
		name := ""
		if len(args) != 0 && len(strings.TrimSpace(args[0])) != 0 {
			name = args[0]
		}

		err = kaytu.DeleteApiKeyRequest(cfg.AccessToken, name)
		if err != nil {
			return err
		}

		fmt.Println("API Key deleted")

		return nil
	},
}
