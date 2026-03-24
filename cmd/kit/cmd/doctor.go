package cmd

import (
	"fmt"
	"net/http"

	"github.com/dunkinfrunkin/kit/internal/config"
	"github.com/dunkinfrunkin/kit/internal/detect"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check kit configuration and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking kit health...")
		fmt.Println()

		creds, err := config.Load()
		if err != nil {
			fmt.Printf("[FAIL] Credentials: %v\n", err)
		} else {
			fmt.Printf("[ OK ] Credentials: %s @ %s\n", creds.Email, creds.Server)

			resp, err := http.Get(creds.Server + "/skills")
			if err != nil {
				fmt.Printf("[FAIL] Server connectivity: %v\n", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode < 400 {
					fmt.Printf("[ OK ] Server connectivity: %s\n", creds.Server)
				} else {
					fmt.Printf("[WARN] Server returned status %d\n", resp.StatusCode)
				}
			}
		}

		tools := detect.DetectTools()
		if len(tools) == 0 {
			fmt.Println("[WARN] No AI tools detected")
		} else {
			for _, t := range tools {
				fmt.Printf("[ OK ] Detected tool: %s\n", t)
			}
		}

		return nil
	},
}
