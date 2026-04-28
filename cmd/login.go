package cmd

import (
	"fmt"

	"insighta-cli/internal/auth"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Insighta Labs+ via GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.Login()
		if err != nil {
			fmt.Printf("❌ Login failed: %v\n", err)
			return
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		err := auth.Logout()
		if err != nil {
			fmt.Printf("❌ Logout failed: %v\n", err)
			return
		}
		fmt.Println("✓ Logged out successfully")
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Run: func(cmd *cobra.Command, args []string) {
		creds, err := auth.LoadCredentials()
		if err != nil {
			fmt.Println("❌ Not logged in. Run 'insighta login' first.")
			return
		}
		fmt.Printf("✓ Logged in as @%s\n", creds.Username)
		fmt.Printf("  Backend: %s\n", creds.BackendURL)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
}
