// Package cmd implements the cosmo CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	endpoint string
)

var rootCmd = &cobra.Command{
	Use:   "cosmo",
	Short: "Cosmonaut CLI — control your data platform from the terminal",
	Long: `cosmo is the command-line interface for Cosmonaut.

It communicates with the Cosmonaut control plane API to manage
components, submit jobs, run queries, and inspect platform state.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&endpoint,
		"endpoint",
		envOr("COSMONAUT_ENDPOINT", "http://localhost:8080"),
		"Cosmonaut control plane endpoint ($COSMONAUT_ENDPOINT)",
	)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(jobCmd)
	rootCmd.AddCommand(queryCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print cosmo version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cosmo v0.1.0")
	},
}
