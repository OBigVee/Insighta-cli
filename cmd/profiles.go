package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"insighta-cli/internal/api"
	"insighta-cli/internal/auth"
	"insighta-cli/internal/display"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
}

// ──────────────────────────────────────────
// profiles list
// ──────────────────────────────────────────

var (
	listGender    string
	listCountry   string
	listAgeGroup  string
	listMinAge    int
	listMaxAge    int
	listSortBy    string
	listOrder     string
	listPage      int
	listLimit     int
)

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles with optional filters",
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching profiles..."
		s.Start()

		params := url.Values{}
		if listGender != "" {
			params.Set("gender", listGender)
		}
		if listCountry != "" {
			params.Set("country_id", listCountry)
		}
		if listAgeGroup != "" {
			params.Set("age_group", listAgeGroup)
		}
		if listMinAge > 0 {
			params.Set("min_age", fmt.Sprintf("%d", listMinAge))
		}
		if listMaxAge > 0 {
			params.Set("max_age", fmt.Sprintf("%d", listMaxAge))
		}
		if listSortBy != "" {
			params.Set("sort_by", listSortBy)
		}
		if listOrder != "" {
			params.Set("order", listOrder)
		}
		if listPage > 0 {
			params.Set("page", fmt.Sprintf("%d", listPage))
		}
		if listLimit > 0 {
			params.Set("limit", fmt.Sprintf("%d", listLimit))
		}

		client := api.NewClient()
		resp, err := client.Get("/api/profiles?" + params.Encode())
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handleErrorResponse(resp)
			return
		}

		var result struct {
			Status     string `json:"status"`
			Page       int    `json:"page"`
			Limit      int    `json:"limit"`
			Total      int    `json:"total"`
			TotalPages int    `json:"total_pages"`
			Data       []map[string]interface{} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", err)
			return
		}

		if len(result.Data) == 0 {
			fmt.Println("No profiles found.")
			return
		}

		headers := []string{"Name", "Gender", "Age", "Age Group", "Country", "Country ID"}
		var rows [][]string
		for _, p := range result.Data {
			rows = append(rows, []string{
				fmt.Sprintf("%v", p["name"]),
				fmt.Sprintf("%v", p["gender"]),
				fmt.Sprintf("%v", p["age"]),
				fmt.Sprintf("%v", p["age_group"]),
				fmt.Sprintf("%v", p["country_name"]),
				fmt.Sprintf("%v", p["country_id"]),
			})
		}

		display.PrintTable(headers, rows)
		fmt.Printf("\nPage %d of %d | Total: %d profiles\n", result.Page, result.TotalPages, result.Total)
	},
}

// ──────────────────────────────────────────
// profiles get <id>
// ──────────────────────────────────────────

var profilesGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a single profile by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Fetching profile..."
		s.Start()

		client := api.NewClient()
		resp, err := client.Get("/api/profiles/" + args[0])
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handleErrorResponse(resp)
			return
		}

		var result struct {
			Status string                 `json:"status"`
			Data   map[string]interface{} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", err)
			return
		}

		p := result.Data
		fmt.Println("┌─────────────────────────────────────────┐")
		fmt.Printf("│ Profile: %-31s │\n", p["name"])
		fmt.Println("├─────────────────────────────────────────┤")
		fmt.Printf("│ ID:          %-27s │\n", p["id"])
		fmt.Printf("│ Gender:      %-27s │\n", p["gender"])
		fmt.Printf("│ Age:         %-27v │\n", p["age"])
		fmt.Printf("│ Age Group:   %-27s │\n", p["age_group"])
		fmt.Printf("│ Country:     %-27s │\n", p["country_name"])
		fmt.Printf("│ Country ID:  %-27s │\n", p["country_id"])
		fmt.Printf("│ Created At:  %-27s │\n", p["created_at"])
		fmt.Println("└─────────────────────────────────────────┘")
	},
}

// ──────────────────────────────────────────
// profiles search "<query>"
// ──────────────────────────────────────────

var profilesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search profiles using natural language",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Searching..."
		s.Start()

		params := url.Values{}
		params.Set("q", args[0])

		client := api.NewClient()
		resp, err := client.Get("/api/profiles/search?" + params.Encode())
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handleErrorResponse(resp)
			return
		}

		var result struct {
			Status     string                   `json:"status"`
			Page       int                      `json:"page"`
			Total      int                      `json:"total"`
			TotalPages int                      `json:"total_pages"`
			Data       []map[string]interface{} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", err)
			return
		}

		if len(result.Data) == 0 {
			fmt.Println("No profiles matched your search.")
			return
		}

		headers := []string{"Name", "Gender", "Age", "Age Group", "Country", "Country ID"}
		var rows [][]string
		for _, p := range result.Data {
			rows = append(rows, []string{
				fmt.Sprintf("%v", p["name"]),
				fmt.Sprintf("%v", p["gender"]),
				fmt.Sprintf("%v", p["age"]),
				fmt.Sprintf("%v", p["age_group"]),
				fmt.Sprintf("%v", p["country_name"]),
				fmt.Sprintf("%v", p["country_id"]),
			})
		}

		display.PrintTable(headers, rows)
		fmt.Printf("\nFound %d profiles (Page %d of %d)\n", result.Total, result.Page, result.TotalPages)
	},
}

// ──────────────────────────────────────────
// profiles create --name "..."
// ──────────────────────────────────────────

var createName string

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile (admin only)",
	Run: func(cmd *cobra.Command, args []string) {
		if createName == "" {
			fmt.Println("❌ --name flag is required")
			return
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Creating profile..."
		s.Start()

		body := fmt.Sprintf(`{"name": "%s"}`, createName)
		client := api.NewClient()
		resp, err := client.Post("/api/profiles", body)
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			handleErrorResponse(resp)
			return
		}

		var result struct {
			Status string                 `json:"status"`
			Data   map[string]interface{} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", err)
			return
		}

		p := result.Data
		fmt.Println("✓ Profile created successfully!")
		fmt.Printf("  Name:    %v\n", p["name"])
		fmt.Printf("  Gender:  %v\n", p["gender"])
		fmt.Printf("  Age:     %v\n", p["age"])
		fmt.Printf("  Country: %v (%v)\n", p["country_name"], p["country_id"])
	},
}

// ──────────────────────────────────────────
// profiles export --format csv
// ──────────────────────────────────────────

var (
	exportFormat   string
	exportGender   string
	exportCountry  string
	exportAgeGroup string
	exportMinAge   int
	exportMaxAge   int
)

var profilesExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export profiles to CSV",
	Run: func(cmd *cobra.Command, args []string) {
		if exportFormat != "csv" {
			fmt.Println("❌ Only csv format is supported. Use --format csv")
			return
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Exporting profiles..."
		s.Start()

		params := url.Values{}
		params.Set("format", "csv")
		if exportGender != "" {
			params.Set("gender", exportGender)
		}
		if exportCountry != "" {
			params.Set("country_id", exportCountry)
		}
		if exportAgeGroup != "" {
			params.Set("age_group", exportAgeGroup)
		}
		if exportMinAge > 0 {
			params.Set("min_age", fmt.Sprintf("%d", exportMinAge))
		}
		if exportMaxAge > 0 {
			params.Set("max_age", fmt.Sprintf("%d", exportMaxAge))
		}

		client := api.NewClient()
		resp, err := client.Get("/api/profiles/export?" + params.Encode())
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handleErrorResponse(resp)
			return
		}

		// Extract filename from Content-Disposition header
		filename := "profiles_export.csv"
		cd := resp.Header.Get("Content-Disposition")
		if cd != "" && strings.Contains(cd, "filename=") {
			parts := strings.Split(cd, "filename=")
			if len(parts) > 1 {
				filename = strings.Trim(parts[1], `"`)
			}
		}

		// Save to current working directory
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("❌ Failed to create file: %v\n", err)
			return
		}
		defer file.Close()

		written, err := io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("❌ Failed to write file: %v\n", err)
			return
		}

		fmt.Printf("✓ Exported to %s (%d bytes)\n", filename, written)
	},
}

// ──────────────────────────────────────────
// Helper: handle error responses
// ──────────────────────────────────────────

func handleErrorResponse(resp *http.Response) {
	var errResp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		fmt.Printf("❌ Error (HTTP %d)\n", resp.StatusCode)
		return
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		fmt.Println("❌ Authentication required. Run 'insighta login' first.")
		// Try to auto-refresh
		creds, err := auth.LoadCredentials()
		if err == nil && creds.RefreshToken != "" {
			fmt.Println("  Attempting token refresh...")
			if auth.RefreshTokens() == nil {
				fmt.Println("  ✓ Token refreshed. Please retry your command.")
			} else {
				fmt.Println("  Token refresh failed. Please run 'insighta login'.")
			}
		}
	case http.StatusForbidden:
		fmt.Printf("❌ Access denied: %s\n", errResp.Message)
	case http.StatusTooManyRequests:
		fmt.Println("❌ Rate limit exceeded. Please wait and try again.")
	default:
		fmt.Printf("❌ Error: %s (HTTP %d)\n", errResp.Message, resp.StatusCode)
	}
}

func init() {
	// List flags
	profilesListCmd.Flags().StringVar(&listGender, "gender", "", "Filter by gender (male/female)")
	profilesListCmd.Flags().StringVar(&listCountry, "country", "", "Filter by country code (e.g., NG)")
	profilesListCmd.Flags().StringVar(&listAgeGroup, "age-group", "", "Filter by age group")
	profilesListCmd.Flags().IntVar(&listMinAge, "min-age", 0, "Minimum age filter")
	profilesListCmd.Flags().IntVar(&listMaxAge, "max-age", 0, "Maximum age filter")
	profilesListCmd.Flags().StringVar(&listSortBy, "sort-by", "", "Sort field (age, gender_probability, created_at)")
	profilesListCmd.Flags().StringVar(&listOrder, "order", "", "Sort order (asc, desc)")
	profilesListCmd.Flags().IntVar(&listPage, "page", 0, "Page number")
	profilesListCmd.Flags().IntVar(&listLimit, "limit", 0, "Items per page")

	// Create flags
	profilesCreateCmd.Flags().StringVar(&createName, "name", "", "Profile name (required)")

	// Export flags
	profilesExportCmd.Flags().StringVar(&exportFormat, "format", "csv", "Export format (csv)")
	profilesExportCmd.Flags().StringVar(&exportGender, "gender", "", "Filter by gender")
	profilesExportCmd.Flags().StringVar(&exportCountry, "country", "", "Filter by country code")
	profilesExportCmd.Flags().StringVar(&exportAgeGroup, "age-group", "", "Filter by age group")
	profilesExportCmd.Flags().IntVar(&exportMinAge, "min-age", 0, "Minimum age filter")
	profilesExportCmd.Flags().IntVar(&exportMaxAge, "max-age", 0, "Maximum age filter")

	// Register subcommands
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesGetCmd)
	profilesCmd.AddCommand(profilesSearchCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesExportCmd)

	rootCmd.AddCommand(profilesCmd)
}
