package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// ServiceImpl implements the Git service interface
type ServiceImpl struct{}

// NewService creates a new Git service
func NewService() Service {
	return &ServiceImpl{}
}

// Clone clones a Git repository to the specified path
func (s *ServiceImpl) Clone(repoPath string, destinationPath string) error {
	var cmd *exec.Cmd

	// If the repo path is the current directory or a relative path
	if repoPath == "." || !filepath.IsAbs(repoPath) {
		// Convert to absolute path
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("failed to resolve repository path: %w", err)
		}

		// Create a local clone (file:// protocol)
		cmd = exec.Command("git", "clone", "file://"+absPath, destinationPath)
	} else {
		// Remote repo or absolute path
		cmd = exec.Command("git", "clone", repoPath, destinationPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w, output: %s", err, output)
	}
	return nil
}

// Checkout checks out the specified branch in the repository
func (s *ServiceImpl) Checkout(repoDirectory string, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = repoDirectory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w, output: %s", branchName, err, output)
	}
	return nil
}
