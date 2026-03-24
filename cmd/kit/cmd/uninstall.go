package cmd

import (
	"fmt"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/dunkinfrunkin/kit/internal/install"
	"github.com/spf13/cobra"
)

var (
	uninstallTarget  string
	uninstallProject bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <ref>",
	Short: "Uninstall skills, hooks, or configs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		opts := install.Options{
			Target:  uninstallTarget,
			Project: uninstallProject,
		}

		namespace, name := parseRef(ref, creds.Email)

		if name != "" {
			return uninstallSingleItem(c, namespace, name, opts)
		}

		return uninstallNamespaceItems(c, namespace, opts)
	},
}

func init() {
	uninstallCmd.Flags().StringVar(&uninstallTarget, "target", "", "target tool (claude, codex, cursor)")
	uninstallCmd.Flags().BoolVar(&uninstallProject, "project", false, "uninstall from project directory")
}

func uninstallSingleItem(c *client.Client, namespace, name string, opts install.Options) error {
	for _, t := range []string{"skill", "hook", "config"} {
		_, err := c.GetItem(namespace, t, name)
		if err != nil {
			continue
		}
		if err := install.Uninstall(t, name, opts); err != nil {
			return err
		}
		if err := trackUninstall(namespace, t, name); err != nil {
			return fmt.Errorf("tracking uninstall: %w", err)
		}
		fmt.Printf("Uninstalled %s %s/%s\n", t, namespace, name)
		return nil
	}
	return fmt.Errorf("item %s/%s not found", namespace, name)
}

func uninstallNamespaceItems(c *client.Client, namespace string, opts install.Options) error {
	for _, t := range []string{"skill", "hook", "config"} {
		items, err := c.ListItems(namespace, t)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := install.Uninstall(t, item.Name, opts); err != nil {
				return err
			}
			if err := trackUninstall(namespace, t, item.Name); err != nil {
				return fmt.Errorf("tracking uninstall: %w", err)
			}
			fmt.Printf("Uninstalled %s %s/%s\n", t, namespace, item.Name)
		}
	}
	return nil
}
