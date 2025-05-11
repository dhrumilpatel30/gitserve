package instance

import (
	"gitserve/internal/models"
	"gitserve/internal/workspace"
)

// Service defines the interface for instance operations
type Service interface {
	// Create creates a new instance
	Create(workspace *workspace.Workspace, branchName string, command string) (*models.Instance, error)

	// RunProcess runs the process and blocks until it completes (for non-detached mode)
	RunProcess(instance *models.Instance) error

	// StartDetachedProcess starts the process in background (for detached mode)
	StartDetachedProcess(instance *models.Instance) error

	// StopProcess stops the process for an instance
	StopProcess(instance *models.Instance) error

	// List returns all instances
	List() ([]*models.Instance, error)

	// Get returns an instance by ID
	Get(id string) (*models.Instance, error)
}
