package utils

import (
	"github.com/spf13/cobra"
	"strconv"
)

func ReadIntFlag(cmd *cobra.Command, name string) int64 {
	str := ReadStringFlag(cmd, name)
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}

func ReadStringFlag(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		return ""
	}
	value := cmd.Flags().Lookup(name).Value.String()
	return value
}

func ReadBooleanFlag(cmd *cobra.Command, name string) bool {
	str := ReadStringFlag(cmd, name)
	i, _ := strconv.ParseBool(str)
	return i
}
