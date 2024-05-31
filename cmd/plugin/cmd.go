package plugin

import (
	"github.com/spf13/cobra"
)

var PluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	PluginCmd.AddCommand(installCmd)
	PluginCmd.AddCommand(uninstallCmd)
	PluginCmd.AddCommand(listCmd)

	installCmd.Flags().String("token", "", "Github fine-grained access token")
	installCmd.Flags().Bool("unsafe", false, "Allow kaytu to install unapproved plugins")
	installCmd.Flags().Bool("plugin-debug-mode", false, "Enable plugin debug mode (manager wont start plugin)")
}
