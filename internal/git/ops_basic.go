package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// runGitCommand executes a git command in the specified directory and returns its combined output.
func (s *ServiceImpl) runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	s.log.Debug("Running git command: git %s (in dir: %s)", strings.Join(args, " "), dir)

	err := cmd.Run()
	stdout := outb.String()
	stderr := errb.String()
	combinedOutput := stdout + stderr // Combine for logging and error messages

	if err != nil {
		s.log.Error("Git command failed: git %s. Error: %v. Output: %s", strings.Join(args, " "), err, combinedOutput)
		return combinedOutput, fmt.Errorf("git %s failed: %w. Output: %s", strings.Join(args, " "), err, combinedOutput)
	}
	s.log.Debug("Git command successful: git %s. Output: %s", strings.Join(args, " "), combinedOutput)
	return combinedOutput, nil
}

// Clone clones a Git repository to the specified path
func (s *ServiceImpl) Clone(repoPath string, destinationPath string) error {
	cloneArgs := []string{}
	effectiveRepoPath := repoPath

	// If the repo path is the current directory or a relative path
	if repoPath == "." || !filepath.IsAbs(repoPath) {
		// Convert to absolute path
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("failed to resolve repository path: %w", err)
		}
		effectiveRepoPath = "file://" + absPath
		// cloneArgs = append(cloneArgs, "clone", "--depth", "1", effectiveRepoPath, destinationPath)
		cloneArgs = append(cloneArgs, "clone", effectiveRepoPath, destinationPath)
	} else {
		// Remote repo or absolute path
		cloneArgs = append(cloneArgs, "clone", effectiveRepoPath, destinationPath)
	}

	_, err := s.runGitCommand("", cloneArgs...) // Run in current dir, git clone creates the destinationPath
	if err != nil {
		// Error is already formatted by runGitCommand, but we add context
		return fmt.Errorf("failed to clone repository '%s' to '%s': %w", repoPath, destinationPath, err)
	}
	s.log.Info("Successfully cloned %s to %s", repoPath, destinationPath)
	return nil
}

// Checkout checks out the specified ref (branch, tag, commit) in the repository
func (s *ServiceImpl) Checkout(repoDirectory string, refName string) error {
	_, err := s.runGitCommand(repoDirectory, "checkout", refName)
	if err != nil {
		return fmt.Errorf("failed to checkout ref '%s': %w", refName, err)
	}
	s.log.Info("Successfully checked out %s in %s", refName, repoDirectory)
	return nil
}
