package workspace

// Workspace represents a temporary workspace
type Workspace struct {
	ID   string
	Path string
}

// Service defines the interface for workspace operations
type Service interface {
	// Create creates a new temporary workspace
	Create() (*Workspace, error)

	// GetPath returns the path for a workspace
	GetPath(workspace *Workspace) string

	// Cleanup removes a workspace
	Cleanup(workspace *Workspace) error
}
