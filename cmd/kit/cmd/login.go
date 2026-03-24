package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dunkinfrunkin/kit/internal/auth"
	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var (
	loginServer   string
	loginSSO      bool
	loginIssuer   string
	loginClientID string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a kit server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginSSO {
			issuer := loginIssuer
			if issuer == "" {
				issuer = os.Getenv("KIT_OIDC_ISSUER")
			}
			clientID := loginClientID
			if clientID == "" {
				clientID = os.Getenv("KIT_OIDC_CLIENT_ID")
			}
			if issuer == "" || clientID == "" {
				return fmt.Errorf("--issuer and --client-id (or KIT_OIDC_ISSUER and KIT_OIDC_CLIENT_ID env vars) are required for SSO login")
			}

			token, email, err := auth.StartPKCEFlow(issuer, clientID)
			if err != nil {
				return fmt.Errorf("SSO login failed: %w", err)
			}

			creds := &config.Credentials{
				Server: loginServer,
				Email:  email,
				Token:  token,
			}
			if err := config.Save(creds); err != nil {
				return err
			}

			fmt.Printf("Logged in as %s (via SSO)\n", email)
			return nil
		}

		fmt.Print("Email: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("no input")
		}
		email := scanner.Text()

		c := client.New(loginServer, "")
		resp, err := c.Login(email)
		if err != nil {
			return err
		}

		creds := &config.Credentials{
			Server: loginServer,
			Email:  resp.Email,
			Token:  resp.Token,
		}
		if err := config.Save(creds); err != nil {
			return err
		}

		fmt.Printf("Logged in as %s\n", resp.Email)
		return nil
	},
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		resp, err := c.CreateToken(name)
		if err != nil {
			return err
		}

		fmt.Printf("Token: %s\n", resp.Token)
		fmt.Println("Warning: this token will only be shown once. Store it securely.")
		return nil
	},
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		tokens, err := c.ListTokens()
		if err != nil {
			return err
		}

		if len(tokens) == 0 {
			fmt.Println("No API tokens.")
			return nil
		}

		for _, t := range tokens {
			fmt.Printf("%-20s  ...%s  %s\n", t.Name, t.Prefix, t.CreatedAt.Format("2006-01-02"))
		}
		return nil
	},
}

var tokenDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Revoke an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		creds, err := config.Load()
		if err != nil {
			return err
		}

		c := client.New(creds.Server, creds.Token)
		if err := c.DeleteToken(name); err != nil {
			return err
		}

		fmt.Printf("Token %q revoked.\n", name)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginServer, "server", "", "server URL")
	loginCmd.MarkFlagRequired("server")
	loginCmd.Flags().BoolVar(&loginSSO, "sso", false, "use SSO/OIDC authentication")
	loginCmd.Flags().StringVar(&loginIssuer, "issuer", "", "OIDC issuer URL")
	loginCmd.Flags().StringVar(&loginClientID, "client-id", "", "OIDC client ID")

	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenDeleteCmd)
}
