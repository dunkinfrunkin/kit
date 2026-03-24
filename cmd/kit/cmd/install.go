package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/dunkinfrunkin/kit/internal/install"
	"github.com/spf13/cobra"
)

var (
	installTarget  string
	installProject bool
	installSkills  bool
	installHooks   bool
	installConfigs bool
	installProfile bool
)

var installCmd = &cobra.Command{
	Use:   "install <ref>",
	Short: "Install skills, hooks, or configs from the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		opts := install.Options{
			Target:  installTarget,
			Project: installProject,
		}

		if installProfile {
			return installProfileItems(c, creds, ref, opts)
		}

		namespace, name := parseRef(ref, creds.Email)

		if name != "" {
			return installSingleItem(c, namespace, name, opts)
		}

		return installNamespaceItems(c, namespace, opts)
	},
}

func init() {
	installCmd.Flags().StringVar(&installTarget, "target", "", "target tool (claude, codex, cursor)")
	installCmd.Flags().BoolVar(&installProject, "project", false, "install to project directory")
	installCmd.Flags().BoolVar(&installSkills, "skills", false, "only install skills")
	installCmd.Flags().BoolVar(&installHooks, "hooks", false, "only install hooks")
	installCmd.Flags().BoolVar(&installConfigs, "configs", false, "only install configs")
	installCmd.Flags().BoolVarP(&installProfile, "profile", "p", false, "install a profile")
}

func parseRef(ref, email string) (namespace, name string) {
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		return parts[0], parts[1]
	}
	return ref, ""
}

func installSingleItem(c *client.Client, namespace, name string, opts install.Options) error {
	for _, t := range []string{"skill", "hook", "config"} {
		item, err := c.GetItem(namespace, t, name)
		if err != nil {
			continue
		}
		return installClientItem(item, opts)
	}
	return fmt.Errorf("item %s/%s not found", namespace, name)
}

func installNamespaceItems(c *client.Client, namespace string, opts install.Options) error {
	types := []string{"skill", "hook", "config"}
	if installSkills || installHooks || installConfigs {
		types = nil
		if installSkills {
			types = append(types, "skill")
		}
		if installHooks {
			types = append(types, "hook")
		}
		if installConfigs {
			types = append(types, "config")
		}
	}

	for _, t := range types {
		items, err := c.ListItems(namespace, t)
		if err != nil {
			return err
		}
		for i := range items {
			full, err := c.GetItem(namespace, t, items[i].Name)
			if err != nil {
				return err
			}
			if err := installClientItem(full, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func installProfileItems(c *client.Client, creds *config.Credentials, ref string, opts install.Options) error {
	namespace, name := parseRef(ref, creds.Email)
	if name == "" {
		name = namespace
		namespace = "@" + creds.Email
	}

	profile, err := c.GetProfile(namespace, name)
	if err != nil {
		return err
	}

	for _, pRef := range profile.Items {
		item, err := c.GetItem(namespace, pRef.Type, pRef.Name)
		if err != nil {
			return fmt.Errorf("fetching %s/%s: %w", pRef.Type, pRef.Name, err)
		}
		if err := installClientItem(item, opts); err != nil {
			return err
		}
	}
	return nil
}

func installClientItem(item *client.Item, opts install.Options) error {
	content, err := base64.StdEncoding.DecodeString(item.Content)
	if err != nil {
		return fmt.Errorf("decoding content for %s: %w", item.Name, err)
	}

	installItem := install.Item{
		Namespace: item.Namespace,
		Type:      item.Type,
		Name:      item.Name,
		Content:   content,
	}

	if err := install.Install(installItem, opts); err != nil {
		return err
	}

	if err := trackInstall(item.Namespace, item.Type, item.Name, item.Version); err != nil {
		return fmt.Errorf("tracking install: %w", err)
	}

	fmt.Printf("Installed %s %s/%s (v%d)\n", item.Type, item.Namespace, item.Name, item.Version)
	return nil
}
