package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/dunkinfrunkin/kit/internal/install"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [namespace]",
	Short: "Update installed items to latest versions",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	installed, err := loadTracking()
	if err != nil {
		return err
	}
	if len(installed) == 0 {
		fmt.Println("Nothing installed.")
		return nil
	}

	var filterNS string
	if len(args) > 0 {
		filterNS = args[0]
	}

	creds, err := config.Load()
	if err != nil {
		return err
	}

	c := client.New(creds.Server, creds.Token)
	updated := 0

	for _, item := range installed {
		if filterNS != "" && item.Namespace != filterNS {
			continue
		}

		latest, err := c.GetItem(item.Namespace, item.Type, item.Name)
		if err != nil {
			fmt.Printf("Warning: could not fetch %s/%s: %v\n", item.Namespace, item.Name, err)
			continue
		}

		if latest.Version <= item.Version {
			continue
		}

		content, err := base64.StdEncoding.DecodeString(latest.Content)
		if err != nil {
			return fmt.Errorf("decoding %s: %w", latest.Name, err)
		}

		installItem := install.Item{
			Namespace: latest.Namespace,
			Type:      latest.Type,
			Name:      latest.Name,
			Content:   content,
		}
		if err := install.Install(installItem, install.Options{}); err != nil {
			return err
		}
		if err := trackInstall(latest.Namespace, latest.Type, latest.Name, latest.Version); err != nil {
			return err
		}

		fmt.Printf("Updated %s %s/%s v%d -> v%d\n", latest.Type, latest.Namespace, latest.Name, item.Version, latest.Version)
		updated++
	}

	if updated == 0 {
		fmt.Println("Everything is up to date.")
	}
	return nil
}
