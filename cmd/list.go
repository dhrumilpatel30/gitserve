package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Git branches",
	Long:  `List all Git branches in the current repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := exec.Command("git", "branch", "-a").Output()
		if err != nil {
			return fmt.Errorf("failed to list branches: %v", err)
		}

		fmt.Print(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
