package cmd

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/flags"
	"github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/cmd/predef"
	awsConfig "github.com/kaytu-io/kaytu/pkg/aws"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kaytu",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := flags.ReadStringFlag(cmd, "profile")
		preferencesFlag := flags.ReadStringFlag(cmd, "preferences")
		if len(preferencesFlag) > 0 {
			cnt, err := os.ReadFile(preferencesFlag)
			if err != nil {
				return err
			}
			var p []preferences.PreferenceItem
			err = yaml.Unmarshal(cnt, &p)
			if err != nil {
				return err
			}
			preferences.DefaultPreferences()
			preferences.Update(p)
		}

		cfg, err := awsConfig.GetConfig(context.Background(), "", "", "", "", &profile, nil)
		if err != nil {
			return err
		}

		client := sts.NewFromConfig(cfg)
		out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
		if err != nil {
			return errors.New("unable to retrieve AWS account details, please check your AWS cli and ensure that you are logged-in.")
		}

		orgClient := organizations.NewFromConfig(cfg)
		orgOut, _ := orgClient.DescribeOrganization(context.Background(), &organizations.DescribeOrganizationInput{})
		config, err := server.GetConfig()
		if err != nil {
			return err
		}

		identification := map[string]string{}
		if config.AccessToken != "" {
			identification["account"] = hash.HashString(*out.Account)
			identification["user_id"] = hash.HashString(*out.UserId)
			identification["sts_arn"] = hash.HashString(*out.Arn)

			if orgOut != nil && orgOut.Organization != nil {
				identification["org_id"] = hash.HashString(*orgOut.Organization.Id)
				identification["org_m_arn"] = hash.HashString(*orgOut.Organization.MasterAccountArn)
				identification["org_m_email"] = hash.HashString(*orgOut.Organization.MasterAccountEmail)
				identification["org_m_account"] = hash.HashString(*orgOut.Organization.MasterAccountId)
			}
		} else {
			identification["account"] = *out.Account
			identification["user_id"] = *out.UserId
			identification["sts_arn"] = *out.Arn

			if orgOut != nil && orgOut.Organization != nil {
				identification["org_id"] = *orgOut.Organization.Id
				identification["org_m_arn"] = *orgOut.Organization.MasterAccountArn
				identification["org_m_email"] = *orgOut.Organization.MasterAccountEmail
				identification["org_m_account"] = *orgOut.Organization.MasterAccountId
			}
		}

		p := tea.NewProgram(view.NewApp(cfg, identification))
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(predef.VersionCmd)
	rootCmd.AddCommand(predef.LoginCmd)
	rootCmd.AddCommand(predef.LogoutCmd)
	rootCmd.Flags().String("profile", "", "AWS profile for authentication")
	rootCmd.Flags().String("preferences", "", "Path to preferences file (yaml)")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
