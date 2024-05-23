package predef

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/api/auth0"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
	"time"
)

const RetrySleep = 3

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Kaytu",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := server.GetConfig()
		if err != nil {
			return err
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
