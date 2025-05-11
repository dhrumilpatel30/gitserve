package runner

import (
	"fmt"
	"gitserve/internal/git"
	"gitserve/internal/instance"
	"gitserve/internal/models"
	"gitserve/internal/validation"
	"gitserve/internal/workspace"
)

// ServiceImpl implements the Runner service interface
type ServiceImpl struct {
	validationService validation.Service
	gitService        git.Service
	workspaceService  workspace.Service
	instanceService   instance.Service
}

// NewService creates a new Runner service
func NewService(
	validationService validation.Service,
	gitService git.Service,
	workspaceService workspace.Service,
	instanceService instance.Service,
) Service {
	return &ServiceImpl{
		validationService: validationService,
		gitService:        gitService,
		workspaceService:  workspaceService,
		instanceService:   instanceService,
	}
}

// Run sets up a Git branch and creates an instance ready to run
func (s *ServiceImpl) Run(request *models.RunRequest) (*models.Instance, error) {
	// Validate the request
	if err := s.validationService.ValidateRunRequest(request); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create a workspace
	ws, err := s.workspaceService.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Get workspace path
	wsPath := s.workspaceService.GetPath(ws)

	// Clone the repository
	repoPath := request.RepoPath
	if repoPath == "" {
		repoPath = "."
	}

	if err := s.gitService.Clone(repoPath, wsPath); err != nil {
		// Cleanup workspace on error
		s.workspaceService.Cleanup(ws)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout the branch
	if err := s.gitService.Checkout(wsPath, request.BranchName); err != nil {
		// Cleanup workspace on error
		s.workspaceService.Cleanup(ws)
		return nil, fmt.Errorf("failed to checkout branch: %w", err)
	}

	// Determine the command to run
	command := request.Command
	if command == "" {
		// Default command if none specified
		command = "npm run dev"
	}

	// Create an instance (but don't start the process yet)
	instance, err := s.instanceService.Create(ws, request.BranchName, command)
	if err != nil {
		// Cleanup workspace on error
		s.workspaceService.Cleanup(ws)
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return instance, nil
}
