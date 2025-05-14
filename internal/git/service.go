package git

import "gitserve/internal/logger"

// ServiceImpl implements the Git service interface
type ServiceImpl struct {
	// githubClient *github.Client // To be initialized, possibly in NewService or lazily
	log logger.Service // Add logger
}

// NewService creates a new Git service
func NewService(log logger.Service) Service { // Add logger parameter
	// TODO: Initialize GitHub client here if needed, e.g.
	// ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	// tc := oauth2.NewClient(context.Background(), ts)
	// client := github.NewClient(tc)
	return &ServiceImpl{log: log /*githubClient: client*/}
}
