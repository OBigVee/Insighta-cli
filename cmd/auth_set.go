package cmd

import (
	"fmt"
	"insighta-cli/internal/auth"

	"github.com/spf13/cobra"
)

var authSetCmd = &cobra.Command{
	Use:   "auth-set",
	Short: "Manually set authentication credentials",
	Long:  "Manually set credentials (useful if browser redirect fails in remote environments)",
	Run: func(cmd *cobra.Command, args []string) {
		accessToken, _ := cmd.Flags().GetString("access")
		refreshToken, _ := cmd.Flags().GetString("refresh")
		username, _ := cmd.Flags().GetString("username")

		if accessToken == "" || refreshToken == "" || username == "" {
			fmt.Println("❌ Error: --access, --refresh, and --username flags are required")
			return
		}

		creds := &auth.Credentials{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Username:     username,
			BackendURL:   "https://stage1.doxantro.com", // Default backend
		}

		err := auth.SaveCredentials(creds)
		if err != nil {
			fmt.Printf("❌ Failed to save credentials: %v\n", err)
			return
		}

		fmt.Printf("✓ Credentials saved manually for @%s\n", username)
	},
}

func init() {
	authSetCmd.Flags().String("access", "", "Access token")
	authSetCmd.Flags().String("refresh", "", "Refresh token")
	authSetCmd.Flags().String("username", "", "GitHub username")
	
	rootCmd.AddCommand(authSetCmd)
}
