package source

import (
	"fmt"
	"regexp"
	"strings"
)

// SourceType represents the type of source for running commands
type SourceType int

const (
	Branch SourceType = iota
	PullRequest
	Commit
	Tag
)

// Source represents a validated source configuration
type Source struct {
	Type   SourceType
	Value  string
	Remote string
}

var (
	// GitHub PR URL patterns
	githubPRRegex = regexp.MustCompile(`^https?://github\.com/[^/]+/[^/]+/pull/\d+/?$`)

	// Commit hash pattern (both full and short forms)
	commitHashRegex = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
)

// ValidateSource ensures the source configuration is valid
func ValidateSource(src *Source) error {
	switch src.Type {
	case Branch:
		if strings.TrimSpace(src.Value) == "" {
			return fmt.Errorf("branch name cannot be empty")
		}
	case PullRequest:
		if !githubPRRegex.MatchString(src.Value) {
			return fmt.Errorf("invalid GitHub PR URL format")
		}
	case Commit:
		if !commitHashRegex.MatchString(strings.ToLower(src.Value)) {
			return fmt.Errorf("invalid commit hash format")
		}
	case Tag:
		if strings.TrimSpace(src.Value) == "" {
			return fmt.Errorf("tag name cannot be empty")
		}
	default:
		return fmt.Errorf("unknown source type")
	}
	return nil
}

// String returns a human-readable representation of the source type
func (st SourceType) String() string {
	switch st {
	case Branch:
		return "branch"
	case PullRequest:
		return "pull request"
	case Commit:
		return "commit"
	case Tag:
		return "tag"
	default:
		return "unknown"
	}
}
