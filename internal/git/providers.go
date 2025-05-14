package git

import "gitserve/internal/models"

// PRProvider defines an interface for provider-specific Pull Request operations.
// This allows for extensibility to support different Git hosting services (GitHub, GitLab, etc.).
type PRProvider interface {
	// GetFetchDetails returns the necessary details for fetching a PR.
	// It should provide:
	// - refSpec: The refspec to use in the `git fetch` command (e.g., "pull/123/head").
	// - localBranchName: The suggested local branch name to check out the PR into (e.g., "pr-123").
	// - error: If any error occurs while determining these details.
	GetFetchDetails(source models.GitSource) (refSpec string, localBranchName string, err error)

	// GetCloneURL (Optional future enhancement)
	// If a PR might need to be cloned from a different repository (e.g., a fork),
	// this method could return the specific URL to clone from.
	// GetCloneURL(source models.GitSource) (string, error)
}
