package git

import (
	// "context" // Will be needed for GitHub client later
	"fmt"
	"gitserve/internal/models"
	"os/exec"
	"strings"
	// _ "github.com/google/go-github/v53/github" // Placeholder, will be added when client is implemented
)

// PrepareRepo clones a repository and checks out the specified source.
func (s *ServiceImpl) PrepareRepo(workspacePath string, source models.GitSource) error {
	s.log.Info("Preparing repository in %s for source type: %s, Ref: %s, Commit: %s, PR: %d",
		workspacePath, source.Type, source.RefName, source.CommitHash, source.PRNumber)

	// Step 1: Clone the base repository
	if err := s.Clone(source.RepoPath, workspacePath); err != nil {
		return fmt.Errorf("initial clone failed for %s: %w", source.RepoPath, err)
	}

	// Step 2: Checkout based on source type
	switch source.Type {
	case models.BranchSource:
		s.log.Info("Checking out branch: %s (Remote: %s)", source.RefName, source.RemoteName)
		// TODO: Handle remote branches if source.RemoteName is not "origin" or if branch is not local after clone.
		if err := s.Checkout(workspacePath, source.RefName); err != nil {
			return err
		}
	case models.CommitSource:
		s.log.Info("Checking out commit: %s", source.CommitHash)
		if err := s.checkoutCommit(workspacePath, source.CommitHash); err != nil {
			return err
		}
	case models.TagSource:
		s.log.Info("Checking out tag: %s", source.RefName)
		if err := s.checkoutTag(workspacePath, source.RefName); err != nil {
			return err
		}
	case models.PRSource:
		s.log.Info("Preparing for PR #%d from %s", source.PRNumber, source.PRApiUrl)

		// Placeholder for GitHub client initialization and use
		// client := s.githubClient // or initialize if nil
		// if client == nil {
		// 	 return fmt.Errorf("GitHub client not initialized for PR operations")
		// }

		// Parse owner/repo from source.RepoPath for GitHub API calls
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(source.RepoPath, "https://github.com/"), ".git"), "/")
		if len(parts) != 2 {
			return fmt.Errorf("could not parse owner/repo from PR base repository URL: %s", source.RepoPath)
		}
		owner, repo := parts[0], parts[1]
		_ = owner // Mark as used for now
		_ = repo  // Mark as used for now

		// TODO: Fetch PR details from GitHub API here to get head repo URL and head branch name.
		// Example: prDetails, err := fetchGitHubPRDetails(client, owner, repo, source.PRNumber)
		// if err != nil { return err }
		// headRepoURL := prDetails.HeadRepoURL
		// headBranch := prDetails.HeadBranch

		// Simplified PR fetch and checkout (assumes PR is from the same repository or directly fetchable)
		prRefSpec := fmt.Sprintf("pull/%d/head", source.PRNumber)
		localBranchName := fmt.Sprintf("pr-%d", source.PRNumber)

		// Use source.RemoteName (defaults to "origin" in cmd/run.go)
		fetchCmd := exec.Command("git", "fetch", source.RemoteName, prRefSpec+":"+localBranchName)
		fetchCmd.Dir = workspacePath
		output, err := fetchCmd.CombinedOutput()
		if err != nil {
			s.log.Error("Failed to fetch PR #%d refspec %s: %s. Output: %s", source.PRNumber, prRefSpec, err, output)
			return fmt.Errorf("failed to fetch PR #%d refspec %s: %w", source.PRNumber, prRefSpec, err)
		}
		s.log.Debug("Fetched PR #%d to local branch %s. Output: %s", source.PRNumber, localBranchName, string(output))

		if err := s.Checkout(workspacePath, localBranchName); err != nil {
			return fmt.Errorf("failed to checkout local PR branch %s for PR #%d: %w", localBranchName, source.PRNumber, err)
		}
		s.log.Info("Successfully checked out PR #%d to branch %s", source.PRNumber, localBranchName)

	case models.UndefinedSource:
		s.log.Error("Cannot prepare repository: git source type is undefined")
		return fmt.Errorf("cannot prepare repository: git source type is undefined")
	default:
		s.log.Error("Unsupported git source type: %s", source.Type.String())
		return fmt.Errorf("unsupported git source type: %s", source.Type.String())
	}

	return nil
}
