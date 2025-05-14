package git

import (
	"fmt"
	// os/exec is no longer needed directly here
)

// checkoutCommit checks out a specific commit hash.
func (s *ServiceImpl) checkoutCommit(repoDirectory string, commitHash string) error {
	output, err := s.runGitCommand(repoDirectory, "checkout", commitHash)
	if err != nil {
		// runGitCommand already logs details, so we just need to contextualize the error
		return fmt.Errorf("failed to checkout commit '%s': %w. Output: %s", commitHash, err, output)
	}
	s.log.Info("Successfully checked out commit %s in %s", commitHash, repoDirectory)
	return nil
}

// checkoutTag checks out a specific tag.
func (s *ServiceImpl) checkoutTag(repoDirectory string, tagName string) error {
	// Using "tags/"+tagName for explicit tag checkout, consistent with previous logic.
	// Git checkout <tagName> often works, but this is more explicit for refs/tags/tagName.
	refToCheckout := "tags/" + tagName
	output, err := s.runGitCommand(repoDirectory, "checkout", refToCheckout)
	if err != nil {
		return fmt.Errorf("failed to checkout tag '%s' (as %s): %w. Output: %s", tagName, refToCheckout, err, output)
	}
	s.log.Info("Successfully checked out tag %s (as %s) in %s", tagName, refToCheckout, repoDirectory)
	return nil
}
