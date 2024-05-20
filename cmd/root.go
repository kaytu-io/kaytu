package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaytu-io/kaytu/cmd/plugin"
	"github.com/kaytu-io/kaytu/cmd/predef"
	"github.com/kaytu-io/kaytu/controller"
	plugin2 "github.com/kaytu-io/kaytu/pkg/plugin"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/kaytu-io/kaytu/pkg/version"
	"github.com/kaytu-io/kaytu/preferences"
	"github.com/kaytu-io/kaytu/view"
	"github.com/rogpeppe/go-internal/semver"
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
	optimizeCmd.PersistentFlags().String("output", "interactive", "Show optimization results in selected output (possible values: interactive, table, csv, json. default value: interactive)")
	optimizeCmd.PersistentFlags().Bool("plugin-debug-mode", false, "Enable plugin debug mode (manager wont start plugin)")
}

func Execute() {
	err := server.CheckForUpdate()
	if err != nil {
		panic(err)
	}

	plugins, err := server.GetPlugins()
	if err != nil {
		panic(err)
	}

	for _, p := range plugins {
		if p.Config.Name == "aws" {
			server.RemoveConfig()
			Execute()
			return
		}
	}

	if len(plugins) == 0 {
		manager := plugin2.New()
		err := manager.StartServer()
		if err != nil {
			panic(err)
		}

		err = manager.Install("aws", "")
		if err != nil {
			panic(err)
		}

		plugins, err = server.GetPlugins()
		if err != nil {
			panic(err)
		}
	}

	for _, loopPlg := range plugins {
		plg := loopPlg
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

					nonInteractiveFlag := utils.ReadStringFlag(c, "output")
					manager := plugin2.New()

					switch nonInteractiveFlag {
					case "interactive":
					case "table":
					case "csv":
					case "json":
					default:
						return fmt.Errorf("output mode not recognized\npossible values: interactive, table, csv, json. default value: interactive (default \"interactive\")")
					}

					if nonInteractiveFlag != "interactive" {
						manager.SetNonInteractiveView()
					}

					pluginDebugMode := utils.ReadBooleanFlag(c, "plugin-debug-mode")
					if pluginDebugMode {
						manager.SetListenPort(30422)
					}

					err = manager.StartServer()
					if err != nil {
						return err
					}

					if !pluginDebugMode {
						repoAddr := "github.com/" + plg.Config.Name
						if plg.Config.Name == "aws" {
							repoAddr = "aws"
						}
						err = manager.Install(repoAddr, "")
						if err != nil {
							fmt.Println("failed due to", err)
						}

						err = manager.StartPlugin(cmd.Name)
						if err != nil {
							return err
						}
					}

					waitLoopCount := 100
					if pluginDebugMode {
						waitLoopCount = 1000
					}

					for i := 0; i < waitLoopCount; i++ {
						runningPlg := manager.GetPlugin(plg.Config.Name)
						if runningPlg != nil {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					runningPlg := manager.GetPlugin(plg.Config.Name)
					if runningPlg == nil {
						return fmt.Errorf("running plugin not found: %s", plg.Config.Name)
					}

					if semver.Compare(version.VERSION, runningPlg.Plugin.Config.MinKaytuVersion) == -1 {
						return fmt.Errorf("plugin requires kaytu version %s, please update your Kaytu CLI", plg.Config.MinKaytuVersion)
					}

					flagValues := map[string]string{}
					for _, flag := range cmd.GetFlags() {
						value := utils.ReadStringFlag(c, flag.Name)
						flagValues[flag.Name] = value
					}

					for _, rcmd := range runningPlg.Plugin.Config.Commands {
						if rcmd.Name == cmd.Name {
							preferences.Update(rcmd.DefaultPreferences)

							if rcmd.LoginRequired && cfg.AccessToken == "" {
								// login
								err := predef.LoginCmd.RunE(c, args)
								if err != nil {
									return err
								}

								cfg, err = server.GetConfig()
								if err != nil {
									return err
								}
							}
							break
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

					if nonInteractiveFlag != "interactive" {
						err := manager.NonInteractiveView.WaitAndShowResults(nonInteractiveFlag)
						return err
					} else {
						helpController := controller.NewHelp()
						helpPage := view.NewHelpPage(helpController)

						jobsController := controller.NewJobs()
						statusBar := view.NewStatusBarView(jobsController, helpController)
						jobsPage := view.NewJobsPage(jobsController)

						optimizationsController := controller.NewOptimizations()
						optimizationsPage := view.NewOptimizationsView(optimizationsController, helpController, statusBar)
						optimizationsDetailsPage := view.NewOptimizationDetailsView(optimizationsController, helpController, statusBar)
						preferencesPage := view.NewPreferencesConfiguration(helpController, optimizationsController, statusBar)

						manager.SetUI(jobsController, optimizationsController)

						p := tea.NewProgram(view.NewApp(
							optimizationsPage,
							optimizationsDetailsPage,
							preferencesPage,
							helpPage,
							jobsPage,
						), tea.WithFPS(10))
						if _, err := p.Run(); err != nil {
							return err
						}

						return nil
					}
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
