package git

import "gitserve/internal/models"

// Service defines the interface for Git operations
type Service interface {
	// Clone clones a Git repository to the specified path
	Clone(repoPath string, destinationPath string) error

	// Checkout checks out the specified branch in the repository
	Checkout(repoDirectory string, branchName string) error

	// PrepareRepo clones a repository and checks out the specified source (branch, commit, tag, or PR)
	PrepareRepo(workspacePath string, source models.GitSource) error
}
