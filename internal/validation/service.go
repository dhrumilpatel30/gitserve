package validation

import (
	"errors"
	"gitserve/internal/models"
	"os"
	"path/filepath"
	"strings"
)

// ServiceImpl implements the Validation service interface
type ServiceImpl struct{}

// NewService creates a new Validation service
func NewService() Service {
	return &ServiceImpl{}
}

// ValidateRunRequest validates a run request
func (s *ServiceImpl) ValidateRunRequest(request *models.RunRequest) error {
	if request.Source.Type == models.UndefinedSource {
		return errors.New("git source type is undefined")
	}

	repoPath := request.Source.RepoPath
	// For PRSource, RepoPath might be derived later or could be the base repo.
	// For other sources, if RepoPath is not explicitly set by CLI parsing, it defaults to current dir.
	if repoPath == "" && request.Source.Type != models.PRSource {
		// Default to current directory if not specified for non-PR sources
		repoPath = "."
	} else if repoPath == "" && request.Source.Type == models.PRSource && request.Source.PRApiUrl == "" {
		// If it's a PR source but RepoPath and PRApiUrl are empty, it's an issue.
		// PRApiUrl is primary for PRs to derive RepoPath.
		return errors.New("PR API URL is required for pull request source when repository path is not derivable")
	}

	// For PRs, RepoPath might be a URL. os.Stat won't work directly on URLs.
	// The actual cloning/access will handle URL validation.
	// For local paths, we perform the Stat check.
	if request.Source.Type != models.PRSource || (request.Source.Type == models.PRSource && !strings.HasPrefix(repoPath, "http")) {
		// Check if the repo path exists (only if it's not a URL)
		_, err := os.Stat(repoPath)
		if err != nil {
			return errors.New("repository path does not exist: " + repoPath)
		}

		// Check if it's a Git repository (only if it's not a URL)
		gitDir := filepath.Join(repoPath, ".git")
		_, gitDirErr := os.Stat(gitDir)
		if gitDirErr != nil {
			return errors.New("not a git repository: " + repoPath)
		}
	}

	switch request.Source.Type {
	case models.BranchSource:
		if request.Source.RefName == "" {
			return errors.New("branch name (RefName) is required for branch source")
		}
	case models.CommitSource:
		if request.Source.CommitHash == "" {
			return errors.New("commit hash is required for commit source")
		}
	case models.TagSource:
		if request.Source.RefName == "" {
			return errors.New("tag name (RefName) is required for tag source")
		}
	case models.PRSource:
		if request.Source.PRNumber <= 0 {
			return errors.New("valid PR number is required for pull request source")
		}
		if request.Source.PRApiUrl == "" { // RepoPath for PRs is often derived from PRApiUrl
			return errors.New("PR API URL is required for pull request source")
		}
		// Further validation of PRApiUrl format could be added here or in a dedicated PR service.
	default:
		return errors.New("unknown git source type")
	}

	return nil
}
