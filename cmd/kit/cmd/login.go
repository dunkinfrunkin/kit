package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var (
	loginServer string
	loginToken  string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a kit server",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := loginToken
		if token == "" {
			fmt.Print("API Token: ")
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				return fmt.Errorf("no input")
			}
			token = strings.TrimSpace(scanner.Text())
		}

		if token == "" {
			return fmt.Errorf("token is required — create one in the dashboard at %s", loginServer)
		}

		c := client.New(loginServer, token)
		email, err := c.Whoami()
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}

		creds := &config.Credentials{
			Server: loginServer,
			Email:  email,
			Token:  token,
		}
		if err := config.Save(creds); err != nil {
			return err
		}

		fmt.Printf("Logged in as %s\n", email)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginServer, "server", "", "server URL")
	loginCmd.MarkFlagRequired("server")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "API token")
}
