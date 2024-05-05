package cmd

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/flags"
	"github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/cmd/predef"
	awsConfig "github.com/kaytu-io/kaytu/pkg/aws"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/kaytu-io/kaytu/pkg/metrics"
	processor2 "github.com/kaytu-io/kaytu/pkg/processor"
	"github.com/kaytu-io/kaytu/pkg/provider"
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
			preferences.DefaultPreferences(map[string]bool{"all": true})
			preferences.Update(p)
		}

		cfg, err := awsConfig.GetConfig(context.Background(), "", "", "", "", &profile, nil)
		if err != nil {
			return err
		}

		prv, err := provider.NewAWS(cfg)
		if err != nil {
			return err
		}

		metricPrv, err := metrics.NewCloudWatch(cfg)
		if err != nil {
			return err
		}

		identification, err := prv.Identify()
		config, err := server.GetConfig()
		if err != nil {
			return err
		}

		if config != nil && config.AccessToken != "" {
			for k, v := range identification {
				identification[k] = hash.HashString(v)
			}
		}

		jobs := view.NewJobsView()
		optimizations := view.NewOptimizationsView()
		ec2Processor := processor2.NewEC2InstanceProcessor(prv, metricPrv, identification, jobs, optimizations)
		rdsProcessor := processor2.NewRDSInstanceProcessor(prv, metricPrv, identification, jobs, optimizations)
		optimizations.SetReEvaluateFunc(ec2Processor.ReEvaluate, rdsProcessor.ReEvaluate)

		p := tea.NewProgram(view.NewApp(optimizations, jobs))
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
