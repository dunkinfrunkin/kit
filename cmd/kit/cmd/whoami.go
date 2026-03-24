package cmd

import (
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Printf("%s @ %s\n", creds.Email, creds.Server)
		return nil
	},
}
