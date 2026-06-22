package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	jobEngine string
	jobFile   string
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage jobs",
}

var jobSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a job to a processing engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jobEngine == "" {
			return fmt.Errorf("--engine is required")
		}
		if jobFile == "" {
			return fmt.Errorf("--file is required")
		}
		// TODO: call POST /api/v1/components/:engine/exec on the control plane
		fmt.Printf("submitting job %q to engine %q via %s\n", jobFile, jobEngine, endpoint)
		return nil
	},
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List running and recent jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: call GET /api/v1/jobs on the control plane
		fmt.Printf("fetching jobs from %s\n", endpoint)
		return nil
	},
}

func init() {
	jobSubmitCmd.Flags().StringVar(&jobEngine, "engine", "", "target engine (spark, flink, ray)")
	jobSubmitCmd.Flags().StringVar(&jobFile, "file", "", "path to the job file")

	jobCmd.AddCommand(jobSubmitCmd)
	jobCmd.AddCommand(jobListCmd)
}
