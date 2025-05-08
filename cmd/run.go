package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [source]",
	Short: "Run a command or server from a branch, pr, or commit",
	Long: `
Run a command or server from different sources. The source can be:
1. A branch name (local or remote)
2. A GitHub PR URL
3. A commit hash (with -C flag)
4. A tag (with -t flag)

Examples:
  gitserve run main                    # Run from main branch
  gitserve run feature/xyz -c "npm i"  # Run command on feature branch
  gitserve run --pr https://github.com/user/repo/pull/123  # Run from PR
  gitserve run --commit abc123def            # Run from commit
  gitserve run --tag v1.0.0               # Run from tag
  gitserve run --port 3000 develop         # Run on port 3000 from develop branch
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(&runOptions.PortNumber, "port", "p", 0, "Port number to run the command on")
	runCmd.Flags().BoolVarP(&runOptions.IsDetached, "detached", "d", false, "Run the command in a detached state")
	runCmd.Flags().StringVarP(&runOptions.CommandToRun, "command", "c", "", "Command to run")
	runCmd.Flags().StringVarP(&runOptions.PRLink, "pr", "r", "", "GitHub PR link")
	runCmd.Flags().StringVarP(&runOptions.BranchName, "branch", "b", "", "Branch name")
	runCmd.Flags().StringVarP(&runOptions.CommitHash, "commit", "C", "", "Commit hash")
	runCmd.Flags().StringVarP(&runOptions.TagName, "tag", "t", "", "Tag name")
	runCmd.Flags().StringVarP(&runOptions.NamedCommand, "name", "n", "", "Named command")
	runCmd.Flags().StringVarP(&runOptions.RemoteName, "remote", "R", "", "Remote name")
}
