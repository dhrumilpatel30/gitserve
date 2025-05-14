package sourceresolver

import "gitserve/internal/models"

// CLIOptions holds the raw command-line options relevant for resolving the Git source.
// This decouples the resolver service from cobra.Command or specific flag structures.
type CLIOptions struct {
	Args       []string // Positional arguments from the command line
	RepoPath   string   // Explicit repository path (optional, often defaults to ".")
	PRLink     string
	BranchName string
	CommitHash string
	TagName    string
	RemoteName string
}

// Service defines the interface for resolving various command-line inputs
// into a structured models.GitSource.
// This helps to keep the cmd package lean and centralizes source determination logic.
type Service interface {
	// Resolve takes the CLI options and determines the appropriate GitSource.
	// It handles defaulting, parsing (like PR URLs), and prioritizing inputs.
	Resolve(opts CLIOptions) (models.GitSource, error)
}
