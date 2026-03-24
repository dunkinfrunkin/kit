package cmd

import (
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:        "uninstall <ref>",
	Short:      "Alias for delete",
	Args:       cobra.ExactArgs(1),
	Hidden:     true,
	RunE:       deleteCmd.RunE,
}
