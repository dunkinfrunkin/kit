package cmd

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update all installed items to latest versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(cmd, nil)
	},
}
