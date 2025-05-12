package instance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		Path:        workspace.Path,
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
	// defer s.mutex.Unlock() // Deferring unlock until after goroutine is spawned if PID is updated within lock

	// Find instance to ensure it exists and get its path
	storedInstance, exists := s.instances[instance.ID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("instance %s not found", instance.ID)
	}

	// The instance.Path should be populated correctly from Create() method
	workspacePath := storedInstance.Path
	if workspacePath == "" {
		// Fallback or error if path somehow not set, though it should be by Create.
		// For now, let's try to get it from s.workspacePaths as a safeguard, though less direct.
		// This part of the logic might be redundant if instance.Path is reliably set.
		var wpExists bool
		workspacePath, wpExists = s.workspacePaths[storedInstance.WorkspaceID]
		if !wpExists {
			s.mutex.Unlock()
			return fmt.Errorf("workspace path for instance %s (ID: %s) not found and not set in instance model", storedInstance.BranchName, instance.ID)
		}
	}
	s.mutex.Unlock() // Unlock before os.Create and cmd.Start to avoid holding lock too long

	// Create cmd
	parts := []string{"sh", "-c", storedInstance.Command}
	cmd := exec.Command(parts[0], parts[1:]...)

	// Set current directory to workspace path
	cmd.Dir = workspacePath

	// Set PGID to enable killing the entire process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Ensure log directory exists (optional, if logs are in a subfolder of workspacePath)
	// For now, logs directly in workspacePath

	// Configure stdout/stderr to be redirected to files within the workspace
	stdoutLogPath := filepath.Join(workspacePath, fmt.Sprintf("%s.out.log", instance.ID))
	stderrLogPath := filepath.Join(workspacePath, fmt.Sprintf("%s.err.log", instance.ID))

	stdoutFile, err := os.Create(stdoutLogPath)
	if err != nil {
		return fmt.Errorf("failed to create stdout log file %s: %w", stdoutLogPath, err)
	}

	stderrFile, err := os.Create(stderrLogPath)
	if err != nil {
		stdoutFile.Close()
		return fmt.Errorf("failed to create stderr log file %s: %w", stderrLogPath, err)
	}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the process
	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Update instance with process ID and status
	// This needs to be done carefully with the lock
	s.mutex.Lock()
	// Re-fetch storedInstance in case of concurrent modifications, though less likely here
	// if si, ok := s.instances[instance.ID]; ok {
	// 	si.ProcessID = cmd.Process.Pid
	// 	si.Status = "running"
	// }
	// Directly update the instance passed in, which is a pointer.
	// Also update the one in the map.
	storedInstance.ProcessID = cmd.Process.Pid
	storedInstance.Status = "running"
	// Update the original instance object that was passed in as well, if it's different
	// from storedInstance (which it is, instance vs storedInstance)
	// The caller (cmd/run.go) has 'instanceModel', which is 'instance' here.
	// So we update 'instance' (which is 'instanceModel' in cmd/run.go)
	instance.ProcessID = cmd.Process.Pid
	instance.Status = "running"
	s.mutex.Unlock()

	// Start a goroutine to wait for the process to complete
	// This is needed to avoid zombie processes
	go func() {
		processErr := cmd.Wait() // Capture error from Wait

		// Close log files
		stdoutFile.Close()
		stderrFile.Close()

		// Update status
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if si, exists := s.instances[instance.ID]; exists {
			if processErr != nil {
				si.Status = "failed" // Or "exited_with_error"
			} else {
				si.Status = "stopped"
			}
			// Also update the original instance object if necessary, though its lifecycle might be over
			instance.Status = si.Status // Reflect final status
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
