package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed items and check for updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		installed, err := loadTracking()
		if err != nil {
			return err
		}
		if len(installed) == 0 {
			fmt.Println("No items installed.")
			return nil
		}

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tTYPE\tNAME\tINSTALLED\tLATEST\tSTATUS")

		for _, item := range installed {
			latest, err := c.GetItem(item.Namespace, item.Type, item.Name)
			status := "up-to-date"
			latestVer := item.Version
			if err != nil {
				status = "unknown"
			} else {
				latestVer = latest.Version
				if latest.Version > item.Version {
					status = "outdated"
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\tv%d\tv%d\t%s\n",
				item.Namespace, item.Type, item.Name, item.Version, latestVer, status)
		}

		w.Flush()
		return nil
	},
}
