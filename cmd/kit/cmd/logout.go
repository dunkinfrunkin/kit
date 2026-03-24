package cmd

import (
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Remove(); err != nil {
			return err
		}
		fmt.Println("Logged out.")
		return nil
	},
}
