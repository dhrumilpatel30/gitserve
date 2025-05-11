package runner

import (
	"gitserve/internal/models"
)

// Service defines the interface for runner operations
type Service interface {
	// Run runs a command from a Git branch
	Run(request *models.RunRequest) (*models.Instance, error)
}
