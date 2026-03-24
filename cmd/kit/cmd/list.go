package cmd

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var listMine bool

var listCmd = &cobra.Command{
	Use:   "list [namespace]",
	Short: "List items",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)

		var namespace string
		if len(args) > 0 {
			namespace = args[0]
		}
		if listMine {
			namespace = "@" + creds.Email
		}

		var allItems []client.Item
		for _, t := range []string{"skill", "hook", "config"} {
			items, err := c.ListItems(namespace, t)
			if err != nil {
				return err
			}
			allItems = append(allItems, items...)
		}

		printItems(allItems)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listMine, "mine", false, "list only your items")
}

func printItems(items []client.Item) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tTYPE\tNAME\tAUTHOR\tVERSION")
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", item.Namespace, item.Type, item.Name, item.Author, item.Version)
	}
	w.Flush()
}
