package predef

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/api/auth0"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
	"time"
)

const RetrySleep = 3

var loginCmd *cobra.Command

// LoginCmd returns the singleton login command
func LoginCmd() *cobra.Command {
	// singletons
	if loginCmd == nil {
		loginCmd = &cobra.Command{
			Use:   "login",
			Short: "Login to Kaytu",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := server.GetConfig()
				if err != nil {
					return err
				}

				authFlag := cmd.Flag("api-key")
				if authFlag != nil {
					authToken := authFlag.Value.String()
					if authToken != "" {
						cfg.AccessToken = authToken
						err = server.SetConfig(*cfg)
						if err != nil {
							return fmt.Errorf("[login-setConfig]: %v", err)
						}
						return nil
					}
				}

				deviceCode, err := auth0.RequestDeviceCode()
				if err != nil {
					return fmt.Errorf("[login-deviceCode]: %v", err)
				}

				var accessToken string
				for i := 0; i < 100; i++ {
					accessToken, err = auth0.AccessToken(deviceCode)
					if err != nil {
						time.Sleep(RetrySleep * time.Second)
						continue
					}
					break
				}
				if err != nil {
					return fmt.Errorf("[login-accessToken]: %v", err)
				}

				cfg.AccessToken = accessToken
				err = server.SetConfig(*cfg)
				if err != nil {
					return fmt.Errorf("[login-setConfig]: %v", err)
				}
				return nil
			},
		}

		loginCmd.Flags().String("api-key", "", "API key - if provided, it'll be set as the access token")
	}
	return loginCmd
}
