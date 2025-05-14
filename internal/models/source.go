package models

// GitSourceType defines the type of the Git source.
// Using iota for enum-like behavior, though direct string constants might also work well.
type GitSourceType int

const (
	UndefinedSource GitSourceType = iota
	BranchSource                  // Represents a branch
	CommitSource                  // Represents a specific commit hash
	TagSource                     // Represents a tag
	PRSource                      // Represents a Pull Request (e.g., from GitHub)
)

// String provides a string representation for GitSourceType - useful for logging/debugging.
func (gst GitSourceType) String() string {
	switch gst {
	case BranchSource:
		return "Branch"
	case CommitSource:
		return "Commit"
	case TagSource:
		return "Tag"
	case PRSource:
		return "PullRequest"
	case UndefinedSource:
		return "Undefined"
	default:
		return "Unknown"
	}
}

// PRProviderType identifies the Git hosting provider for a Pull Request.
type PRProviderType int

const (
	UndefinedProvider PRProviderType = iota
	GitHubProvider
	// GitLabProvider // Example for future extension
	// BitbucketProvider // Example for future extension
)

// String provides a string representation for PRProviderType.
func (ppt PRProviderType) String() string {
	switch ppt {
	case GitHubProvider:
		return "GitHub"
	// case GitLabProvider:
	// 	return "GitLab"
	// case BitbucketProvider:
	// 	return "Bitbucket"
	case UndefinedProvider:
		return "UndefinedProvider"
	default:
		return "UnknownProvider"
	}
}

// GitSource specifies the details of the Git entity to be checked out.
// This structure will be populated based on CLI arguments.
type GitSource struct {
	Type GitSourceType

	// RepoPath specifies the primary repository URL or local path to clone from.
	// For PRs, this would be the base repository URL.
	RepoPath string

	// RefName is the primary reference: branch name, tag name.
	// For commits, CommitHash is used directly.
	// For PRs, this might be the target branch of the PR or the head branch name fetched locally.
	RefName string

	CommitHash string // Specific commit SHA to checkout.
	PRNumber   int    // Pull Request number (e.g., for GitHub).
	RemoteName string // Optional: name of the remote (e.g., "origin", "upstream"). Defaults to "origin".

	// PRProvider indicates the source control provider for PRSource type.
	PRProvider PRProviderType

	// For PRs, these might be populated after fetching PR details from an API:
	PRApiUrl      string // Full URL to the PR (e.g. GitHub PR URL provided by user)
	PRHeadRepoURL string // Clone URL of the repository containing the PR's head branch.
	PRHeadBranch  string // Name of the head branch in the PR's source repository.
	PRBaseRepoURL string // Clone URL of the repository the PR targets.
	PRBaseBranch  string // Name of the base branch in the PR's target repository.

}
