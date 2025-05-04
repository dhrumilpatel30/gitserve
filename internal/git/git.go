package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BranchManager handles Git branch operations
type BranchManager struct {
	repoPath string
	workDir  string
}

// NewBranchManager creates a new BranchManager instance
func NewBranchManager(repoPath, workDir string) *BranchManager {
	return &BranchManager{
		repoPath: repoPath,
		workDir:  workDir,
	}
}

// CloneBranch clones a specific branch into the working directory
func (bm *BranchManager) CloneBranch(branchName string) error {
	// Clean up existing directory if it exists
	if err := bm.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup existing directory: %v", err)
	}

	// Ensure the working directory exists
	if err := os.MkdirAll(bm.workDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %v", err)
	}

	// Get the absolute path of the repository
	absRepoPath, err := filepath.Abs(bm.repoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute repo path: %v", err)
	}

	// Verify the branch exists before cloning
	if err := bm.verifyBranchExists(branchName); err != nil {
		return err
	}

	// Clone the repository with the specific branch
	cmd := exec.Command("git", "clone", "-b", branchName, "--single-branch", absRepoPath, bm.workDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone branch: %v\nOutput: %s", err, output)
	}

	return nil
}

// verifyBranchExists checks if the branch exists in the repository
func (bm *BranchManager) verifyBranchExists(branchName string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	cmd.Dir = bm.repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}
	return nil
}

// Cleanup removes the working directory
func (bm *BranchManager) Cleanup() error {
	// Check if directory exists before attempting to remove
	if _, err := os.Stat(bm.workDir); err == nil {
		if err := os.RemoveAll(bm.workDir); err != nil {
			return fmt.Errorf("failed to remove directory: %v", err)
		}
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch
func (bm *BranchManager) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = bm.workDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return string(output), nil
}
