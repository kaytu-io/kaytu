package plugin

import (
	"errors"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin"
	"github.com/kaytu-io/kaytu/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

var installCmd = &cobra.Command{
	Use: "install",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		pluginDebugMode := utils.ReadBooleanFlag(cmd, "plugin-debug-mode")

		manager := plugin.New()
		if pluginDebugMode {
			manager.SetListenPort(30422)
		}

		err := manager.StartServer()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return errors.New("please provide plugin path")
		}

		token := utils.ReadStringFlag(cmd, "token")
		unsafe := utils.ReadBooleanFlag(cmd, "unsafe")
		err = manager.Install(ctx, args[0], token, unsafe, pluginDebugMode)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("failed to install plugin due to %v\n", err))
			return err
		}
		return nil
	},
}
