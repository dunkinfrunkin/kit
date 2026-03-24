package cmd

import (
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var profileTeam string

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		namespace := "@" + creds.Email
		if profileTeam != "" {
			namespace = profileTeam
		}

		c := client.New(creds.Server, creds.Token)
		if err := c.CreateProfile(namespace, name); err != nil {
			return err
		}

		fmt.Printf("Created profile %s/%s\n", namespace, name)
		return nil
	},
}

var profileAddCmd = &cobra.Command{
	Use:   "add <profile> <item-ref>",
	Short: "Add an item to a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		itemRef := args[1]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		namespace := "@" + creds.Email
		if profileTeam != "" {
			namespace = profileTeam
		}

		c := client.New(creds.Server, creds.Token)

		itemNS, itemName := parseRef(itemRef, creds.Email)
		if itemName == "" {
			itemName = itemNS
			itemNS = namespace
		}

		var itemType string
		for _, t := range []string{"skill", "hook", "config"} {
			if _, err := c.GetItem(itemNS, t, itemName); err == nil {
				itemType = t
				break
			}
		}
		if itemType == "" {
			return fmt.Errorf("item %s not found", itemRef)
		}

		ref := client.ProfileRef{
			Name: itemName,
			Type: itemType,
		}
		if err := c.AddProfileItem(namespace, profileName, ref); err != nil {
			return err
		}

		fmt.Printf("Added %s %s to profile %s/%s\n", itemType, itemName, namespace, profileName)
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list <name>",
	Short: "List items in a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		namespace := "@" + creds.Email
		if profileTeam != "" {
			namespace = profileTeam
		}

		c := client.New(creds.Server, creds.Token)
		profile, err := c.GetProfile(namespace, name)
		if err != nil {
			return err
		}

		fmt.Printf("Profile: %s/%s\n", profile.Namespace, profile.Name)
		for _, item := range profile.Items {
			fmt.Printf("  %s  %s\n", item.Type, item.Name)
		}
		return nil
	},
}

func init() {
	profileCmd.PersistentFlags().StringVar(&profileTeam, "team", "", "team namespace")
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileListCmd)
}
