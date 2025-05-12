package cmd

import (
	"fmt"
	"gitserve/internal/git"
	"gitserve/internal/instance"
	"gitserve/internal/models"
	"gitserve/internal/runner"
	"gitserve/internal/storage"
	"gitserve/internal/validation"
	"gitserve/internal/workspace"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// ANSI color codes (can be moved to a shared package later)
const (
	colorResetRun = "\033[0m"
	colorGreenRun = "\033[32m"
	colorBoldRun  = "\033[1m"
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

		// Storage service for instances
		storeDataPath := filepath.Join(homeDir, ".gitserve", "store")
		instanceStore, err := storage.NewJSONInstanceStore(storeDataPath)
		if err != nil {
			return fmt.Errorf("failed to initialize instance store: %w", err)
		}

		// Create the runner service
		runnerService := runner.NewService(
			validationService,
			gitService,
			workspaceService,
			instanceService,
		)

		// Run the command (prepares the instance)
		instanceModel, err := runnerService.Run(request) // instanceModel is *models.Instance
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		fmt.Printf("Instance %s%s%s prepared for branch %s%s%s (Path: %s)\n",
			colorBoldRun, instanceModel.ID, colorResetRun,
			colorBoldRun, instanceModel.BranchName, colorResetRun,
			instanceModel.Path)

		// Handle based on detached mode
		if runOptions.IsDetached {
			// In detached mode, start the process in background
			if err := instanceService.StartDetachedProcess(instanceModel); err != nil { // This updates instanceModel.PID, Status
				return fmt.Errorf("failed to start detached process: %w", err)
			}
			fmt.Printf("%sProcess is running in detached mode.%s\n", colorGreenRun, colorResetRun)

			// Add to instance store
			storageInst := storage.Instance{
				ID:         instanceModel.ID,
				Name:       fmt.Sprintf("%s-%s", instanceModel.BranchName, instanceModel.ID[:8]), // Generate a user-friendly name
				PID:        instanceModel.ProcessID,
				Port:       instanceModel.Port,                                                             // Ensure Port is correctly populated if used
				Path:       instanceModel.Path,                                                             // This now comes from models.Instance
				Status:     instanceModel.Status,                                                           // Should be "running"
				StartTime:  time.Now().UTC(),                                                               // Record start time
				LogPath:    filepath.Join(instanceModel.Path, fmt.Sprintf("%s.out.log", instanceModel.ID)), // Base log path, actual files are .out.log and .err.log
				GitServeID: "",                                                                             // Placeholder or version info
				// Consider adding BranchName explicitly to storage.Instance if needed for querying
			}

			if err := instanceStore.AddInstance(storageInst); err != nil {
				// Log this error, but maybe don't fail the whole run command?
				// Or decide if this is critical. For now, let's report it.
				return fmt.Errorf("failed to save instance to store (process is running): %w", err)
			}
			fmt.Printf("%sInstance %s%s%s (PGID: %s%d%s) details saved. Use 'gitserve list'.%s\n",
				colorGreenRun,
				colorBoldRun, storageInst.ID, colorResetRun,
				colorBoldRun, storageInst.PID, colorResetRun,
				colorResetRun)

			return nil
		} else {
			// In non-detached mode, run the process directly
			fmt.Printf("%sProcess is running in foreground. Press Ctrl+C to stop.%s\n", colorYellow, colorResetRun) // Assuming colorYellow is defined or add it
			if err := instanceService.RunProcess(instanceModel); err != nil {
				// Just log the error but don't fail - process exiting with error code is expected
				_, _ = fmt.Fprintf(os.Stderr, "%sProcess exited with error: %v%s\n", colorRed, err, colorResetRun) // Assuming colorRed is defined
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
