package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "insighta",
	Short: "Insighta Labs+ CLI — Profile Intelligence Platform",
	Long:  `A CLI tool for interacting with the Insighta Labs+ Profile Intelligence System.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
