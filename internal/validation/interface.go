package validation

import (
	"gitserve/internal/models"
)

// Service defines the interface for validation operations
type Service interface {
	// ValidateRunRequest validates a run request
	ValidateRunRequest(request *models.RunRequest) error
}
