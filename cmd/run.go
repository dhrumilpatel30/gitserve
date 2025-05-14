package cmd

import (
	"fmt"
	"gitserve/internal/git"
	"gitserve/internal/instance"
	"gitserve/internal/logger"
	"gitserve/internal/models"
	"gitserve/internal/runner"
	"gitserve/internal/storage"
	"gitserve/internal/validation"
	"gitserve/internal/workspace"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
		log := logger.NewService(logger.LogLevelInfo)

		var gitSource models.GitSource
		gitSource.RepoPath = "."
		if runOptions.RemoteName != "" {
			gitSource.RemoteName = runOptions.RemoteName
		} else {
			gitSource.RemoteName = "origin"
		}

		if runOptions.PRLink != "" {
			gitSource.Type = models.PRSource
			gitSource.PRApiUrl = runOptions.PRLink
			parsedURL, err := url.Parse(runOptions.PRLink)
			if err != nil {
				return fmt.Errorf("invalid PR URL %s: %w", runOptions.PRLink, err)
			}
			pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
			if parsedURL.Hostname() == "github.com" && len(pathParts) == 4 && pathParts[2] == "pull" {
				owner := pathParts[0]
				repo := pathParts[1]
				prNumStr := pathParts[3]
				prNum, convErr := fmt.Sscan(prNumStr, &gitSource.PRNumber)
				if convErr != nil || prNum == 0 {
					return fmt.Errorf("could not parse PR number from URL %s: %v", runOptions.PRLink, convErr)
				}
				gitSource.RepoPath = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
			} else {
				return fmt.Errorf("unsupported PR URL format or host: %s. Only GitHub PRs are currently supported", runOptions.PRLink)
			}
		} else if runOptions.CommitHash != "" {
			gitSource.Type = models.CommitSource
			gitSource.CommitHash = runOptions.CommitHash
		} else if runOptions.TagName != "" {
			gitSource.Type = models.TagSource
			gitSource.RefName = runOptions.TagName
		} else if runOptions.BranchName != "" {
			gitSource.Type = models.BranchSource
			gitSource.RefName = runOptions.BranchName
		} else if len(args) > 0 {
			gitSource.Type = models.BranchSource
			gitSource.RefName = args[0]
		} else {
			gitSource.Type = models.BranchSource
			gitSource.RefName = "main"
			log.Warning("No specific source provided, defaulting to branch 'main'.")
		}

		request := &models.RunRequest{
			Source:   gitSource,
			Detached: runOptions.IsDetached,
			Command:  runOptions.CommandToRun,
		}

		validationService := validation.NewService()
		gitService := git.NewService(log)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		workspacesDir := filepath.Join(homeDir, ".gitserve", "workspaces")
		workspaceService := workspace.NewService(workspacesDir)
		instanceService := instance.NewService()
		storeDataPath := filepath.Join(homeDir, ".gitserve", "store")
		instanceStore, err := storage.NewJSONInstanceStore(storeDataPath)
		if err != nil {
			return fmt.Errorf("failed to initialize instance store: %w", err)
		}
		runnerService := runner.NewService(
			validationService,
			gitService,
			workspaceService,
			instanceService,
			instanceStore,
			log,
		)

		finalInstanceModel, err := runnerService.Run(request)
		if err != nil {
			instanceIDForError := "unknown"
			branchNameForError := "unknown"
			statusForError := "unknown"
			if finalInstanceModel != nil {
				instanceIDForError = finalInstanceModel.ID
				branchNameForError = finalInstanceModel.BranchName
				statusForError = finalInstanceModel.Status
			}
			log.Error("Run failed for instance %s (Ref: %s, Status: %s): %v",
				instanceIDForError, branchNameForError, statusForError, err)
			return err
		}

		if request.Detached {
			log.Info("Instance %s (Ref: %s, PID: %d) is running detached and saved.",
				finalInstanceModel.ID, finalInstanceModel.BranchName, finalInstanceModel.ProcessID)
			log.Info("Workspace: %s. Use 'gitserve list' and 'gitserve logs %s'.",
				finalInstanceModel.Path, finalInstanceModel.ID)
		} else {
			log.Info("Foreground process for instance %s (Ref: %s) completed with status: %s.",
				finalInstanceModel.ID, finalInstanceModel.BranchName, finalInstanceModel.Status)
			log.Info("Workspace %s cleaned up.", finalInstanceModel.Path)
		}
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
