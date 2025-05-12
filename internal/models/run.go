package models

// RunRequest represents the parameters for running a Git branch
type RunRequest struct {
	BranchName string
	RepoPath   string
	Detached   bool
	Command    string // Command to run
}

// Instance represents a running instance of a Git branch
type Instance struct {
	ID          string
	BranchName  string
	WorkspaceID string
	Path        string // Filesystem path of the cloned repository in the workspace
	ProcessID   int
	Port        int
	Status      string
	Command     string
}

type RunOptions struct {
}
