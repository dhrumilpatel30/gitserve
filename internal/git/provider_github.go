package git

import (
	"fmt"
	"gitserve/internal/models"
	// "strings" // Not needed here if all string manipulation is for RepoPath parsing, which stays in pr_handler for now
)

// GitHubProvider implements the PRProvider interface for GitHub.
type GitHubProvider struct{}

// NewGitHubProvider creates a new GitHubProvider.
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{}
}

// GetFetchDetails returns the Git refspec and local branch name for fetching a GitHub PR.
// For GitHub, this is typically "pull/<PR_NUMBER>/head" and "pr-<PR_NUMBER>".
func (ghp *GitHubProvider) GetFetchDetails(source models.GitSource) (refSpec string, localBranchName string, err error) {
	if source.PRNumber <= 0 {
		return "", "", fmt.Errorf("invalid PR number for GitHub PR: %d", source.PRNumber)
	}
	refSpec = fmt.Sprintf("pull/%d/head", source.PRNumber)
	localBranchName = fmt.Sprintf("pr-%d", source.PRNumber)
	return refSpec, localBranchName, nil
}
