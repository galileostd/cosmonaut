package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage registered components",
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered components and their health status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: call GET /api/v1/components on the control plane
		fmt.Printf("fetching components from %s\n", endpoint)
		return nil
	},
}

var componentGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get details of a specific component",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: call GET /api/v1/components/:name on the control plane
		fmt.Printf("fetching component %q from %s\n", args[0], endpoint)
		return nil
	},
}

func init() {
	componentCmd.AddCommand(componentListCmd)
	componentCmd.AddCommand(componentGetCmd)
}
