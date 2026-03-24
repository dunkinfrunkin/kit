package cmd

import (
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <ref>",
	Short: "Delete an item from the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		namespace, name := parseRef(ref, creds.Email)

		if name == "" {
			return fmt.Errorf("specify an item to delete, e.g. %s/item-name", namespace)
		}

		for _, t := range []string{"skill", "hook", "config"} {
			_, err := c.GetItem(namespace, t, name)
			if err != nil {
				continue
			}
			if err := c.DeleteItem(namespace, t, name); err != nil {
				return err
			}
			fmt.Printf("Deleted %s %s/%s from server\n", t, namespace, name)
			return nil
		}
		return fmt.Errorf("item %s/%s not found on server", namespace, name)
	},
}
