package cmd

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/plugin"
	"github.com/kaytu-io/kaytu/cmd/predef"
	plugin2 "github.com/kaytu-io/kaytu/pkg/plugin"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/kaytu-io/kaytu/view"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

var optimizeCmd = &cobra.Command{
	Use: "optimize",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kaytu",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(plugin.PluginCmd)
	rootCmd.AddCommand(predef.VersionCmd)
	rootCmd.AddCommand(predef.LoginCmd)
	rootCmd.AddCommand(predef.LogoutCmd)
	rootCmd.AddCommand(optimizeCmd)

	optimizeCmd.PersistentFlags().String("preferences", "", "Path to preferences file (yaml)")
}

func Execute() {
	plugins, err := server.GetPlugins()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, plg := range plugins {
		for _, loopCmd := range plg.Config.Commands {
			cmd := loopCmd
			theCmd := &cobra.Command{
				Use:  cmd.Name,
				Long: cmd.Description,
				RunE: func(c *cobra.Command, args []string) error {
					cfg, err := server.GetConfig()
					if err != nil {
						return err
					}

					manager := plugin2.New()
					err = manager.StartServer()
					if err != nil {
						return err
					}

					err = manager.StartPlugin(cmd.Name)
					if err != nil {
						return err
					}

					for i := 0; i < 100; i++ {
						runningPlg := manager.GetPlugin(plg.Config.Name)
						if runningPlg != nil {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					runningPlg := manager.GetPlugin(plg.Config.Name)
					if runningPlg == nil {
						return errors.New("running plugin not found")
					}

					flagValues := map[string]string{}
					for _, flag := range cmd.GetFlags() {
						value := utils.ReadStringFlag(c, flag.Name)
						flagValues[flag.Name] = value
					}

					for _, rcmd := range runningPlg.Plugin.Config.Commands {
						if rcmd.Name == cmd.Name {
							preferences.Update(rcmd.DefaultPreferences)
						}
					}

					preferencesFlag := utils.ReadStringFlag(c, "preferences")
					if len(preferencesFlag) > 0 {
						cnt, err := os.ReadFile(preferencesFlag)
						if err != nil {
							return err
						}
						var p []*golang.PreferenceItem
						err = yaml.Unmarshal(cnt, &p)
						if err != nil {
							return err
						}
						preferences.Update(p)
					}

					err = runningPlg.Stream.Send(&golang.ServerMessage{
						ServerMessage: &golang.ServerMessage_Start{
							Start: &golang.StartProcess{
								Command:          cmd.Name,
								Flags:            flagValues,
								KaytuAccessToken: cfg.AccessToken,
							},
						},
					})
					if err != nil {
						return err
					}

					jobs := view.NewJobsView()
					optimizations := view.NewOptimizationsView()
					manager.SetUI(jobs, optimizations)

					p := tea.NewProgram(view.NewApp(optimizations, jobs))
					if _, err := p.Run(); err != nil {
						return err
					}

					return nil
				},
			}

			optimizeCmd.AddCommand(theCmd)
			for _, flag := range cmd.Flags {
				theCmd.Flags().String(flag.Name, flag.Default, flag.Description)
				if flag.Required {
					cobra.MarkFlagRequired(theCmd.Flags(), flag.Name)
				}
			}
		}
	}

	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
