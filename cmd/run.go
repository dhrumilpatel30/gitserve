package cmd

import (
	"fmt"
	"gitserve/internal/git"
	"gitserve/internal/instance"
	"gitserve/internal/models"
	"gitserve/internal/runner"
	"gitserve/internal/validation"
	"gitserve/internal/workspace"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var runOptions struct {
	PortNumber   int
	IsDetached   bool
	CommandToRun string
	PRLink       string
	BranchName   string
	CommitHash   string
	TagName      string
	NamedCommand string
	RemoteName   string
}

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
		// Determine the branch name
		branchName := runOptions.BranchName
		// If branch name flag not set but positional arg exists, use that
		if branchName == "" && len(args) > 0 {
			branchName = args[0]
		}

		// Create the run request
		request := &models.RunRequest{
			BranchName: branchName,
			RepoPath:   ".", // Default to current directory
			Detached:   runOptions.IsDetached,
			Command:    runOptions.CommandToRun,
		}

		// Initialize services
		validationService := validation.NewService()
		gitService := git.NewService()

		// Create a temp directory for workspaces
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		workspacesDir := filepath.Join(homeDir, ".gitserve", "workspaces")
		workspaceService := workspace.NewService(workspacesDir)

		// Instance service
		instanceService := instance.NewService()

		// Create the runner service
		runnerService := runner.NewService(
			validationService,
			gitService,
			workspaceService,
			instanceService,
		)

		// Run the command
		instance, err := runnerService.Run(request)
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		fmt.Printf("Started instance %s for branch %s\n", instance.ID, instance.BranchName)

		// Handle based on detached mode
		if runOptions.IsDetached {
			// In detached mode, start the process in background
			if err := instanceService.StartDetachedProcess(instance); err != nil {
				return fmt.Errorf("failed to start detached process: %w", err)
			}
			fmt.Println("Process is running in detached mode. Use 'gitserve list' to view instances.")
			return nil
		} else {
			// In non-detached mode, run the process directly
			fmt.Println("Process is running. Press Ctrl+C to stop.")
			if err := instanceService.RunProcess(instance); err != nil {
				// Just log the error but don't fail - process exiting with error code is expected
				_, err := fmt.Fprintf(os.Stderr, "Process exited with error: %v\n", err)
				if err != nil {
					return err
				}
			}
			return nil
		}
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
