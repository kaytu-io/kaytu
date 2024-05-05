package utils

import "github.com/spf13/cobra"

func ReadStringFlag(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		return ""
	}
	value := cmd.Flags().Lookup(name).Value.String()
	return value
}
