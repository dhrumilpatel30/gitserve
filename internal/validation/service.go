package validation

import (
	"errors"
	"gitserve/internal/models"
	"os"
	"path/filepath"
)

// ServiceImpl implements the Validation service interface
type ServiceImpl struct{}

// NewService creates a new Validation service
func NewService() Service {
	return &ServiceImpl{}
}

// ValidateRunRequest validates a run request
func (s *ServiceImpl) ValidateRunRequest(request *models.RunRequest) error {
	if request.BranchName == "" {
		return errors.New("branch name is required")
	}

	repoPath := request.RepoPath
	if repoPath == "" {
		// Default to current directory if not specified
		repoPath = "."
	}

	// Check if the repo path exists
	_, err := os.Stat(repoPath)
	if err != nil {
		return errors.New("repository path does not exist")
	}

	// Check if it's a Git repository
	gitDir := filepath.Join(repoPath, ".git")
	_, err = os.Stat(gitDir)
	if err != nil {
		return errors.New("not a git repository")
	}

	return nil
}
