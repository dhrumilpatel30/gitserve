package instance

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"gitserve/internal/models"
	"gitserve/internal/workspace"

	"github.com/google/uuid"
)

// ServiceImpl implements the Instance service interface
type ServiceImpl struct {
	instances      map[string]*models.Instance
	workspacePaths map[string]string // Map workspace IDs to workspace paths
	mutex          sync.RWMutex
}

// NewService creates a new Instance service
func NewService() Service {
	return &ServiceImpl{
		instances:      make(map[string]*models.Instance),
		workspacePaths: make(map[string]string),
	}
}

// Create creates a new instance
func (s *ServiceImpl) Create(workspace *workspace.Workspace, branchName string, command string) (*models.Instance, error) {
	id := uuid.New().String()

	instance := &models.Instance{
		ID:          id,
		BranchName:  branchName,
		WorkspaceID: workspace.ID,
		Status:      "created",
		Command:     command,
	}

	s.mutex.Lock()
	s.instances[id] = instance
	s.workspacePaths[workspace.ID] = workspace.Path // Store the workspace path
	s.mutex.Unlock()

	return instance, nil
}

// RunProcess runs the process for an instance - blocks until the process completes
func (s *ServiceImpl) RunProcess(instance *models.Instance) error {
	s.mutex.Lock()
	// Find instance to ensure it exists
	storedInstance, exists := s.instances[instance.ID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("instance %s not found", instance.ID)
	}

	// Get the workspace path
	workspacePath, exists := s.workspacePaths[storedInstance.WorkspaceID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("workspace path for instance %s not found", instance.ID)
	}
	s.mutex.Unlock()

	// Create cmd
	parts := []string{"sh", "-c", storedInstance.Command}
	cmd := exec.Command(parts[0], parts[1:]...)

	// Set current directory to workspace path
	cmd.Dir = workspacePath

	// Configure stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Update status
	s.mutex.Lock()
	storedInstance.Status = "running"
	s.mutex.Unlock()

	// Run the command (this blocks until it completes)
	err := cmd.Run()

	// Update status when done
	s.mutex.Lock()
	storedInstance.Status = "stopped"
	s.mutex.Unlock()

	// If it's an exit error, that's expected (non-zero exit code)
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("process exited with code %d", exitErr.ExitCode())
	}

	return err
}

// StartDetachedProcess starts the process in the background for an instance
func (s *ServiceImpl) StartDetachedProcess(instance *models.Instance) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find instance to ensure it exists
	storedInstance, exists := s.instances[instance.ID]
	if !exists {
		return fmt.Errorf("instance %s not found", instance.ID)
	}

	// Get the workspace path
	workspacePath, exists := s.workspacePaths[storedInstance.WorkspaceID]
	if !exists {
		return fmt.Errorf("workspace path for instance %s not found", instance.ID)
	}

	// Create cmd
	parts := []string{"sh", "-c", storedInstance.Command}
	cmd := exec.Command(parts[0], parts[1:]...)

	// Set current directory to workspace path
	cmd.Dir = workspacePath

	// Configure stdout/stderr to be redirected to files
	stdoutFile, err := os.Create(fmt.Sprintf("%s.out.log", instance.ID))
	if err != nil {
		return fmt.Errorf("failed to create stdout log file: %w", err)
	}

	stderrFile, err := os.Create(fmt.Sprintf("%s.err.log", instance.ID))
	if err != nil {
		stdoutFile.Close()
		return fmt.Errorf("failed to create stderr log file: %w", err)
	}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the process
	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Update instance with process ID
	storedInstance.ProcessID = cmd.Process.Pid
	storedInstance.Status = "running"

	// Start a goroutine to wait for the process to complete
	// This is needed to avoid zombie processes
	go func() {
		cmd.Wait()

		// Close log files
		stdoutFile.Close()
		stderrFile.Close()

		// Update status
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if stored, exists := s.instances[instance.ID]; exists {
			stored.Status = "stopped"
		}
	}()

	return nil
}

// StopProcess stops the process for an instance
func (s *ServiceImpl) StopProcess(instance *models.Instance) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	storedInstance, exists := s.instances[instance.ID]
	if !exists {
		return fmt.Errorf("instance %s not found", instance.ID)
	}

	if storedInstance.Status != "running" {
		return fmt.Errorf("instance %s is not running", instance.ID)
	}

	proc, err := os.FindProcess(storedInstance.ProcessID)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send a SIGTERM signal
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}

	storedInstance.Status = "stopping"

	return nil
}

// List returns all instances
func (s *ServiceImpl) List() ([]*models.Instance, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instances := make([]*models.Instance, 0, len(s.instances))
	for _, instance := range s.instances {
		instances = append(instances, instance)
	}

	return instances, nil
}

// Get returns an instance by ID
func (s *ServiceImpl) Get(id string) (*models.Instance, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	instance, exists := s.instances[id]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", id)
	}

	return instance, nil
}
