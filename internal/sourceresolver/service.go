package sourceresolver

import (
	"fmt"
	"gitserve/internal/logger" // Assuming a logger might be useful here
	"gitserve/internal/models"
	"net/url"
	"strings"
)

// ServiceImpl implements the sourceresolver.Service interface.
type ServiceImpl struct {
	log logger.Service // Optional: if logging is needed during resolution
}

// NewService creates a new ServiceImpl for resolving Git sources.
func NewService(log logger.Service) Service {
	return &ServiceImpl{log: log}
}

// Resolve takes CLI options and determines the GitSource.
// This logic is moved from cmd/run.go to centralize source determination.
func (s *ServiceImpl) Resolve(opts CLIOptions) (models.GitSource, error) {
	var gitSource models.GitSource

	// Default RepoPath if not provided, can be overridden by PR parsing.
	gitSource.RepoPath = "." // Default to current directory
	if opts.RepoPath != "" {
		gitSource.RepoPath = opts.RepoPath
	}

	if opts.RemoteName != "" {
		gitSource.RemoteName = opts.RemoteName
	} else {
		gitSource.RemoteName = "origin" // Default remote name
	}

	if opts.PRLink != "" {
		gitSource.Type = models.PRSource
		gitSource.PRApiUrl = opts.PRLink
		parsedURL, err := url.Parse(opts.PRLink)
		if err != nil {
			return models.GitSource{}, fmt.Errorf("invalid PR URL %s: %w", opts.PRLink, err)
		}
		pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

		// Determine PR provider based on hostname
		switch parsedURL.Hostname() {
		case "github.com":
			gitSource.PRProvider = models.GitHubProvider
			if len(pathParts) == 4 && pathParts[2] == "pull" {
				owner := pathParts[0]
				repo := pathParts[1]
				prNumStr := pathParts[3]
				prNum, convErr := fmt.Sscan(prNumStr, &gitSource.PRNumber)
				if convErr != nil || prNum == 0 {
					return models.GitSource{}, fmt.Errorf("could not parse GitHub PR number from URL %s: %v", opts.PRLink, convErr)
				}
				gitSource.RepoPath = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
			} else {
				return models.GitSource{}, fmt.Errorf("invalid GitHub PR URL format: %s", opts.PRLink)
			}
		// case "gitlab.com": // Example for future GitLab support
		// 	 gitSource.PRProvider = models.GitLabProvider
		// 	 // Add GitLab specific parsing logic here
		// 	 return models.GitSource{}, fmt.Errorf("GitLab PRs are not yet supported")
		default:
			gitSource.PRProvider = models.UndefinedProvider
			return models.GitSource{}, fmt.Errorf("unsupported PR provider for URL %s. Only GitHub is currently supported", opts.PRLink)
		}
	} else if opts.CommitHash != "" {
		gitSource.Type = models.CommitSource
		gitSource.CommitHash = opts.CommitHash
	} else if opts.TagName != "" {
		gitSource.Type = models.TagSource
		gitSource.RefName = opts.TagName
	} else if opts.BranchName != "" {
		gitSource.Type = models.BranchSource
		gitSource.RefName = opts.BranchName
	} else if len(opts.Args) > 0 && opts.Args[0] != "" {
		gitSource.Type = models.BranchSource
		gitSource.RefName = opts.Args[0] // First positional argument is branch name
	} else {
		// Default to main/master branch if nothing else specified
		gitSource.Type = models.BranchSource
		gitSource.RefName = "main" // Or "master", consider making this configurable or auto-detected
		s.log.Warning("No specific source (branch, tag, commit, PR) provided, defaulting to branch 'main'.")
	}

	if gitSource.Type == models.UndefinedSource {
		// This case should ideally be caught by the logic above, but as a safeguard:
		return models.GitSource{}, fmt.Errorf("could not determine git source from the provided options")
	}

	s.log.Info("Resolved Git source: Type=%s, RepoPath=%s, RefName=%s, CommitHash=%s, PRNumber=%d, Remote=%s, Provider=%s",
		gitSource.Type, gitSource.RepoPath, gitSource.RefName, gitSource.CommitHash, gitSource.PRNumber, gitSource.RemoteName, gitSource.PRProvider)
	return gitSource, nil
}
