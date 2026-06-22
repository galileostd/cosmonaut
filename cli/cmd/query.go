package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var queryEngine string

var queryCmd = &cobra.Command{
	Use:   "query <sql>",
	Short: "Run a SQL query via the configured query engine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: call POST /api/v1/components/:engine/exec on the control plane
		fmt.Printf("running query on engine %q via %s\n", queryEngine, endpoint)
		fmt.Printf("SQL: %s\n", args[0])
		return nil
	},
}

func init() {
	queryCmd.Flags().StringVar(&queryEngine, "engine", "trino", "query engine to use (default: trino)")
}
