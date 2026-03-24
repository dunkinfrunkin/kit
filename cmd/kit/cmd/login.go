package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dunkinfrunkin/kit/internal/client"
	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/spf13/cobra"
)

var loginServer string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a kit server",
	RunE: func(cmd *cobra.Command, args []string) error {
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

func init() {
	loginCmd.Flags().StringVar(&loginServer, "server", "", "server URL")
	loginCmd.MarkFlagRequired("server")
}
