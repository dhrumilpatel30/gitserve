package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gitserve/internal/git"
	"gitserve/internal/server"

	"github.com/spf13/cobra"
)

var (
	port     int
	detached bool
	command  string
)

var runCmd = &cobra.Command{
	Use:   "run [branch]",
	Short: "Run a server from a specific Git branch",
	Long: `Run a server or command from a specific Git branch in an isolated environment.
The command will:
1. Clone the specified branch into a temporary directory
2. Run any setup commands (e.g., installing dependencies)
3. Start the server or run the specified command`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]

		// Get current working directory as repo path
		repoPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}

		// Create a unique working directory for this branch
		workDir := filepath.Join(os.TempDir(), "gitserve", branchName)

		// Initialize git manager
		gitManager := git.NewBranchManager(repoPath, workDir)

		// Clone the branch
		if err := gitManager.CloneBranch(branchName); err != nil {
			return fmt.Errorf("failed to clone branch: %v", err)
		}

		// Create and start the server
		srv := server.NewServer(branchName, port, workDir, command, detached)

		if err := srv.Start(); err != nil {
			// Cleanup on error
			err := gitManager.Cleanup()
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to start server: %v", err)
		}

		if detached {
			fmt.Printf("Server started in detached mode from branch '%s'\n", branchName)
			fmt.Printf("Working directory: %s\n", workDir)
			fmt.Printf("Port: %d\n", port)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVarP(&port, "port", "p", 3000, "Port to run the server on")
	runCmd.Flags().BoolVarP(&detached, "detach", "d", false, "Run server in detached mode")
	runCmd.Flags().StringVarP(&command, "command", "c", "npm start", "Command to run in the branch directory")
}
