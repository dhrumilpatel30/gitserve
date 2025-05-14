package git

import (
	// "context" // Will be needed for GitHub client later
	"fmt"
	"gitserve/internal/models"
	// "os/exec" // No longer needed directly
	// _ "github.com/google/go-github/v53/github" // Placeholder, will be added when client is implemented
)

// PrepareRepo clones a repository and checks out the specified source.
func (s *ServiceImpl) PrepareRepo(workspacePath string, source models.GitSource) error {
	s.log.Info("Preparing repository in %s for source type: %s, Ref: %s, Commit: %s, PR: %d",
		workspacePath, source.Type, source.RefName, source.CommitHash, source.PRNumber)

	// Step 1: Clone the base repository
	if err := s.Clone(source.RepoPath, workspacePath); err != nil {
		// s.Clone now uses runGitCommand, which logs errors.
		// The error returned by s.Clone is already contextualized.
		return err // No need to wrap further, already: fmt.Errorf("failed to clone repository '%s' to '%s': %w", repoPath, destinationPath, err)
	}

	// Step 2: Checkout based on source type
	switch source.Type {
	case models.BranchSource:
		s.log.Info("Checking out branch: %s (Remote: %s)", source.RefName, source.RemoteName)
		// TODO: Handle remote branches if source.RemoteName is not "origin" or if branch is not local after clone.
		// Current s.Checkout should handle this fine if the branch exists locally or can be fast-forwarded from remote tracking.
		if err := s.Checkout(workspacePath, source.RefName); err != nil {
			return err // s.Checkout now uses runGitCommand and returns contextualized error
		}
	case models.CommitSource:
		s.log.Info("Checking out commit: %s", source.CommitHash)
		if err := s.checkoutCommit(workspacePath, source.CommitHash); err != nil {
			return err // s.checkoutCommit now uses runGitCommand and returns contextualized error
		}
	case models.TagSource:
		s.log.Info("Checking out tag: %s", source.RefName)
		if err := s.checkoutTag(workspacePath, source.RefName); err != nil {
			return err // s.checkoutTag now uses runGitCommand and returns contextualized error
		}
	case models.PRSource:
		// Delegate to the dedicated PR handler
		if err := s.handlePRSource(workspacePath, source); err != nil {
			return err // handlePRSource will return a contextualized error
		}
	case models.UndefinedSource:
		s.log.Error("Cannot prepare repository: git source type is undefined")
		return fmt.Errorf("cannot prepare repository: git source type is undefined")
	default:
		s.log.Error("Unsupported git source type: %s", source.Type.String())
		return fmt.Errorf("unsupported git source type: %s", source.Type.String())
	}

	return nil
}
