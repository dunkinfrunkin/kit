package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var (
	pushTeam     string
	pushItemType string
)

var pushCmd = &cobra.Command{
	Use:   "push <path>",
	Short: "Push a skill, hook, or config to the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		content, itemType, err := readAndDetect(path, pushItemType)
		if err != nil {
			return err
		}

		namespace := "@" + creds.Email
		if pushTeam != "" {
			namespace = pushTeam
		}

		name := itemName(path, itemType)

		c := client.New(creds.Server, creds.Token)
		if err := c.PushItem(namespace, itemType, name, content); err != nil {
			return err
		}

		fmt.Printf("Pushed %s %s/%s\n", itemType, namespace, name)
		return nil
	},
}

func init() {
	pushCmd.Flags().StringVar(&pushTeam, "team", "", "team namespace")
	pushCmd.Flags().StringVar(&pushItemType, "type", "", "item type (skill, hook, config)")
}

func readAndDetect(path, forceType string) ([]byte, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}

	if info.IsDir() {
		skillPath := filepath.Join(path, "SKILL.md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			return nil, "", fmt.Errorf("directory does not contain SKILL.md")
		}
		t := "skill"
		if forceType != "" {
			t = forceType
		}
		return content, t, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	if forceType != "" {
		return content, forceType, nil
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js":
		return content, "hook", nil
	case ".md":
		return content, "config", nil
	default:
		return nil, "", fmt.Errorf("cannot auto-detect type for %q; use --type", ext)
	}
}

func itemName(path string, itemType string) string {
	info, _ := os.Stat(path)
	if info != nil && info.IsDir() {
		return filepath.Base(path)
	}
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
