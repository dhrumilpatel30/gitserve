package git

import (
	"fmt"
	"gitserve/internal/models"
	"strings"
)

// handlePRSource manages the specifics of preparing a repository for a Pull Request source.
// It uses the PRProvider field from the source to delegate to the correct provider implementation.
func (s *ServiceImpl) handlePRSource(workspacePath string, source models.GitSource) error {
	var provider PRProvider

	s.log.Info("Handling PR #%d from %s (Provider Type: %s)", source.PRNumber, source.PRApiUrl, source.PRProvider)

	switch source.PRProvider {
	case models.GitHubProvider:
		provider = NewGitHubProvider() // In a more complex setup, this might come from a map in ServiceImpl
	// case models.GitLabProvider:
	// 	 provider = NewGitLabProvider() // Example
	case models.UndefinedProvider:
		return fmt.Errorf("cannot handle PR from undefined provider for URL: %s", source.PRApiUrl)
	default:
		return fmt.Errorf("unsupported PR provider type: %s for URL: %s", source.PRProvider, source.PRApiUrl)
	}

	refSpec, localBranchName, err := provider.GetFetchDetails(source)
	if err != nil {
		return fmt.Errorf("failed to get fetch details for PR #%d from %s: %w", source.PRNumber, source.PRProvider, err)
	}

	s.log.Info("Preparing PR #%d using %s provider (RefSpec: %s, LocalBranch: %s)",
		source.PRNumber, source.PRProvider, refSpec, localBranchName)

	// The RepoPath parsing for owner/repo could also be moved to the provider if it's provider-specific
	// for API interactions (which are future work).
	if source.PRProvider == models.GitHubProvider && strings.HasPrefix(source.RepoPath, "https://github.com/") {
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(source.RepoPath, "https://github.com/"), ".git"), "/")
		if len(parts) != 2 {
			s.log.Warning("Could not parse owner/repo from GitHub PR base repository URL: %s. Proceeding with direct fetch.", source.RepoPath)
		} else {
			s.log.Debug("Parsed GitHub owner/repo for potential API use: %s/%s", parts[0], parts[1])
		}
	}

	// Use source.RemoteName (defaults to "origin" in cmd/run.go)
	fetchArgs := []string{"fetch", source.RemoteName, refSpec + ":" + localBranchName}
	fetchOutput, err := s.runGitCommand(workspacePath, fetchArgs...)
	if err != nil {
		return fmt.Errorf("failed to fetch PR #%d (refspec %s, provider %s): %w. Output: %s",
			source.PRNumber, refSpec, source.PRProvider, err, fetchOutput)
	}
	s.log.Debug("Fetched PR #%d to local branch %s. Output: %s", source.PRNumber, localBranchName, fetchOutput)

	if err := s.Checkout(workspacePath, localBranchName); err != nil {
		return fmt.Errorf("failed to checkout local PR branch %s for PR #%d (provider %s): %w",
			localBranchName, source.PRNumber, source.PRProvider, err)
	}
	s.log.Info("Successfully checked out PR #%d (provider %s) to branch %s",
		source.PRNumber, source.PRProvider, localBranchName)
	return nil
}
