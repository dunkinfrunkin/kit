package cmd

import (
	"fmt"
	"strings"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <namespace/name>",
	Short: "Show details about an item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		parts := strings.SplitN(ref, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("ref must be in the form namespace/name")
		}
		namespace, name := parts[0], parts[1]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)

		for _, t := range []string{"skill", "hook", "config"} {
			item, err := c.GetItem(namespace, t, name)
			if err != nil {
				continue
			}
			fmt.Printf("Namespace: %s\n", item.Namespace)
			fmt.Printf("Type:      %s\n", item.Type)
			fmt.Printf("Name:      %s\n", item.Name)
			fmt.Printf("Author:    %s\n", item.Author)
			fmt.Printf("Version:   %d\n", item.Version)
			return nil
		}

		return fmt.Errorf("item %s not found", ref)
	},
}
