package git

import (
	"fmt"
	"os/exec"
)

// checkoutCommit checks out a specific commit hash.
func (s *ServiceImpl) checkoutCommit(repoDirectory string, commitHash string) error {
	cmd := exec.Command("git", "checkout", commitHash)
	cmd.Dir = repoDirectory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout commit '%s': %w, output: %s", commitHash, err, output)
	}
	return nil
}

// checkoutTag checks out a specific tag.
func (s *ServiceImpl) checkoutTag(repoDirectory string, tagName string) error {
	// Tags are typically checked out as 'tags/tagName' or just tagName if it creates a detached HEAD
	// or a local branch tracking the tag.
	// For simplicity, let's checkout the tag directly, which results in a detached HEAD.
	cmd := exec.Command("git", "checkout", "tags/"+tagName)
	cmd.Dir = repoDirectory
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout tag '%s': %w, output: %s", tagName, err, output)
	}
	return nil
}
