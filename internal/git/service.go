package git

import (
	"gitserve/internal/logger"
	"gitserve/internal/models"
)

// ServiceImpl implements the Git service interface
type ServiceImpl struct {
	// githubClient *github.Client // To be initialized, possibly in NewService or lazily
	log         logger.Service
	prProviders map[models.PRProviderType]PRProvider // Map to hold PR provider implementations
}

// NewService creates a new Git service
func NewService(log logger.Service) Service { // Add logger parameter
	// TODO: Initialize GitHub client here if needed for providers
	// ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	// tc := oauth2.NewClient(context.Background(), ts)
	// client := github.NewClient(tc)

	serviceImpl := &ServiceImpl{
		log:         log,
		prProviders: make(map[models.PRProviderType]PRProvider),
		// githubClient: client // If client is initialized
	}

	// Register known PR providers
	// When a real GitHub client is added, it would be passed to NewGitHubProvider if needed.
	serviceImpl.prProviders[models.GitHubProvider] = NewGitHubProvider()
	// Example for future: serviceImpl.prProviders[models.GitLabProvider] = NewGitLabProvider(log, gitlabClient)

	return serviceImpl
}
