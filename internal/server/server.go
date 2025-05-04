package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Server represents a running server instance
type Server struct {
	ID       string
	Branch   string
	Port     int
	WorkDir  string
	Command  string
	Process  *os.Process
	Detached bool
}

// NewServer creates a new Server instance
func NewServer(branch string, port int, workDir string, command string, detached bool) *Server {
	return &Server{
		ID:       filepath.Base(workDir), // Using the directory name as ID
		Branch:   branch,
		Port:     port,
		WorkDir:  workDir,
		Command:  command,
		Detached: detached,
	}
}

// installDependencies attempts to install project dependencies
func (s *Server) installDependencies() error {
	// Check if package.json exists
	if _, err := os.Stat(filepath.Join(s.WorkDir, "package.json")); err == nil {
		fmt.Println("Installing npm dependencies...")
		cmd := exec.Command("npm", "install")
		cmd.Dir = s.WorkDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %v", err)
		}
		fmt.Println("Dependencies installed successfully")
	}

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	// Install dependencies first
	if err := s.installDependencies(); err != nil {
		return err
	}

	// Split the command into program and arguments
	args := []string{"sh", "-c", s.Command}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = s.WorkDir

	// Set environment variables
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", s.Port))

	// If not detached, we want to see the output
	if !s.Detached {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// Set up process group for better process management
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.Process = cmd.Process

	if !s.Detached {
		// Wait for the process if not detached
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("server process ended with error: %v", err)
		}
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.Process == nil {
		return fmt.Errorf("no process to stop")
	}

	// Kill the entire process group
	pgid, err := syscall.Getpgid(s.Process.Pid)
	if err != nil {
		return fmt.Errorf("failed to get process group: %v", err)
	}

	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	return nil
}

// IsRunning checks if the server is running
func (s *Server) IsRunning() bool {
	if s.Process == nil {
		return false
	}

	// Try to send signal 0 to check if process exists
	return s.Process.Signal(syscall.Signal(0)) == nil
}
