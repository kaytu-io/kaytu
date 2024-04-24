package cmd

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/flags"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/cmd/predef"
	awsConfig "github.com/kaytu-io/kaytu/pkg/aws"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kaytu",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := flags.ReadStringFlag(cmd, "profile")
		cfg, err := awsConfig.GetConfig(context.Background(), "", "", "", "", &profile, nil)
		if err != nil {
			return err
		}

		client := sts.NewFromConfig(cfg)
		out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
		if err != nil {
			return err
		}
		accountHash := hash.HashString(*out.Account)
		userIdHash := hash.HashString(*out.UserId)
		arnHash := hash.HashString(*out.Arn)

		p := tea.NewProgram(view.NewApp(cfg, accountHash, userIdHash, arnHash))
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(predef.VersionCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
