package runner

import (
	"fmt"
	// "gitserve/internal/git" // No longer directly using gitService.Clone or gitService.Checkout here
	"gitserve/internal/git" // Ensuring git.Service is available for PrepareRepo
	"gitserve/internal/instance"
	"gitserve/internal/logger" // Import logger
	"gitserve/internal/models"
	"gitserve/internal/storage"
	"gitserve/internal/validation"
	"gitserve/internal/workspace"
	"path/filepath"
	"time"
)

// ServiceImpl implements the Runner service interface
type ServiceImpl struct {
	validationService validation.Service
	gitService        git.Service
	workspaceService  workspace.Service
	instanceService   instance.Service
	instanceStore     storage.InstanceStore
	log               logger.Service // Add logger to struct
}

// NewService creates a new Runner service
func NewService(
	validationService validation.Service,
	gitService git.Service,
	workspaceService workspace.Service,
	instanceService instance.Service,
	instanceStore storage.InstanceStore,
	log logger.Service, // Add logger to parameters
) Service {
	return &ServiceImpl{
		validationService: validationService,
		gitService:        gitService,
		workspaceService:  workspaceService,
		instanceService:   instanceService,
		instanceStore:     instanceStore,
		log:               log, // Initialize logger
	}
}

// Run sets up a Git source, executes the command, and manages instance state.
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
	wsPath := s.workspaceService.GetPath(ws)

	// --- Modified Git Setup ---
	s.log.Info("Preparing repository in workspace: %s", wsPath)
	if err := s.gitService.PrepareRepo(wsPath, request.Source); err != nil {
		s.workspaceService.Cleanup(ws)
		// Error already logged by gitService.PrepareRepo if it uses the logger
		return nil, fmt.Errorf("failed to prepare repository from source (%s %s): %w",
			request.Source.Type, request.Source.RefName, err)
	}
	s.log.Info("Repository prepared successfully.")
	// --- End Modified Git Setup ---

	// Determine the command to run (will use config later)
	command := request.Command
	if command == "" {
		command = "npm run dev" // Placeholder
	}

	// Create an instance model
	// The BranchName for models.Instance should reflect the primary reference being worked on.
	// For branches and tags, this is request.Source.RefName.
	// For commits, it could be request.Source.CommitHash (or a shortened version).
	// For PRs, it could be the local PR branch name (e.g., "pr-123") that git.Service checks out.
	// For now, let's use a general approach; instance.Service might need adjustment or a clear contract.
	var instanceRefName string
	switch request.Source.Type {
	case models.BranchSource, models.TagSource:
		instanceRefName = request.Source.RefName
	case models.CommitSource:
		instanceRefName = request.Source.CommitHash
		if len(instanceRefName) > 12 { // Shorten commit hash for display name
			instanceRefName = instanceRefName[:12]
		}
	case models.PRSource:
		instanceRefName = fmt.Sprintf("pr-%d", request.Source.PRNumber)
	default:
		instanceRefName = "unknown-ref"
	}

	instanceModel, err := s.instanceService.Create(ws, instanceRefName, command)
	if err != nil {
		s.workspaceService.Cleanup(ws)
		return nil, fmt.Errorf("failed to create instance model: %w", err)
	}

	// Execute the command based on detached mode (logic remains largely the same)
	if request.Detached {
		s.log.Info("Starting process in detached mode for instance %s (Ref: %s)...", instanceModel.ID, instanceRefName)
		if err := s.instanceService.StartDetachedProcess(instanceModel); err != nil {
			s.workspaceService.Cleanup(ws)
			// s.log.Error already handled by the caller (cmd/run.go) which has access to finalInstanceModel
			return instanceModel, fmt.Errorf("failed to start detached process: %w", err)
		}
		storageInst := storage.Instance{
			ID:         instanceModel.ID,
			Name:       fmt.Sprintf("%s-%s", instanceModel.BranchName, instanceModel.ID[:8]), // BranchName is now more generic ref name
			PID:        instanceModel.ProcessID,
			Port:       instanceModel.Port,
			Path:       instanceModel.Path,
			Status:     instanceModel.Status,
			StartTime:  time.Now().UTC(),
			LogPath:    filepath.Join(instanceModel.Path, fmt.Sprintf("%s.out.log", instanceModel.ID)),
			GitServeID: "",
		}
		if err := s.instanceStore.AddInstance(storageInst); err != nil {
			// s.log.Error might be appropriate here, but caller also handles it.
			return instanceModel, fmt.Errorf("failed to save instance to store: %w", err)
		}
		s.log.Info("Instance %s (PID: %d, Ref: %s) is running in detached mode. Logs: %s",
			instanceModel.ID, instanceModel.ProcessID, instanceModel.BranchName, storageInst.LogPath)
		return instanceModel, nil
	} else {
		s.log.Info("Process is running in foreground for instance %s (Ref: %s). Press Ctrl+C to stop.", instanceModel.ID, instanceRefName)
		runErr := s.instanceService.RunProcess(instanceModel)
		if runErr != nil {
			// Replace fmt.Fprintf with s.log.Error (or Warning depending on if runErr is a true error or just non-zero exit)
			// For now, using Error, assuming runErr implies a problem beyond just non-zero exit.
			s.log.Error("Foreground process for instance %s (Ref: %s) exited: %v", instanceModel.ID, instanceModel.BranchName, runErr)
			return instanceModel, fmt.Errorf("foreground process error: %w", runErr)
		}
		s.log.Info("Foreground process for instance %s (Ref: %s) completed.", instanceModel.ID, instanceModel.BranchName)
		s.workspaceService.Cleanup(ws)
		return instanceModel, nil
	}
}
