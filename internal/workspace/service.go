package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ServiceImpl implements the Workspace service interface
type ServiceImpl struct {
	BaseDir string
}

// NewService creates a new Workspace service
func NewService(baseDir string) Service {
	// Create the base directory if it doesn't exist
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0755)
	}
	return &ServiceImpl{
		BaseDir: baseDir,
	}
}

// Create creates a new temporary workspace
func (s *ServiceImpl) Create() (*Workspace, error) {
	id := uuid.New().String()
	path := filepath.Join(s.BaseDir, id)

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return &Workspace{
		ID:   id,
		Path: path,
	}, nil
}

// GetPath returns the path for a workspace
func (s *ServiceImpl) GetPath(workspace *Workspace) string {
	return workspace.Path
}

// Cleanup removes a workspace
func (s *ServiceImpl) Cleanup(workspace *Workspace) error {
	if err := os.RemoveAll(workspace.Path); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}
	return nil
}
