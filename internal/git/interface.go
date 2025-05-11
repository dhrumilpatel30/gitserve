package git

// Service defines the interface for Git operations
type Service interface {
	// Clone clones a Git repository to the specified path
	Clone(repoPath string, destinationPath string) error

	// Checkout checks out the specified branch in the repository
	Checkout(repoDirectory string, branchName string) error
}
