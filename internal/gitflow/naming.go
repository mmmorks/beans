package gitflow

import (
	"fmt"
	"regexp"
	"strings"
)

// beanIDPattern matches bean IDs in branch names (e.g., "beans-abc123" or just "abc123").
var beanIDPattern = regexp.MustCompile(`^([a-z]+-)?([a-z0-9]+)`)

// BuildBranchName creates a branch name from bean ID and slug.
// Format: {bean-id}/{slug} or just {bean-id} if slug is empty.
// Example: "beans-abc123/user-authentication" or "beans-abc123"
func BuildBranchName(beanID, slug string) string {
	if slug == "" {
		return beanID
	}
	sanitized := SanitizeSlug(slug)
	return fmt.Sprintf("%s/%s", beanID, sanitized)
}

// ParseBranchName extracts the bean ID from a branch name.
// Returns the bean ID and true if successfully parsed, or empty string and false otherwise.
// Handles both formats: "beans-abc123/slug" and "beans-abc123"
func ParseBranchName(branchName string) (beanID string, ok bool) {
	// Try to match the bean ID pattern at the start
	matches := beanIDPattern.FindStringSubmatch(branchName)
	if len(matches) < 3 {
		return "", false
	}

	// Reconstruct the full ID (prefix + id part)
	if matches[1] != "" {
		// Has prefix (e.g., "beans-")
		beanID = matches[1][:len(matches[1])-1] + "-" + matches[2]
	} else {
		// No prefix, just the ID
		beanID = matches[2]
	}

	// Validate it looks like a bean ID (has at least the ID part)
	if matches[2] == "" {
		return "", false
	}

	return beanID, true
}

// SanitizeSlug ensures a slug is git-branch safe.
// Converts to lowercase, replaces invalid characters with hyphens,
// and removes consecutive hyphens.
func SanitizeSlug(slug string) string {
	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Replace any non-alphanumeric characters (except hyphens) with hyphens
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "-")

	// Remove consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length to 50 characters (reasonable for branch names)
	if len(slug) > 50 {
		slug = slug[:50]
		// Trim any trailing hyphen after truncation
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// containsBranchName checks if a commit message contains a reference to the branch name.
// Used for detecting merge commits.
func containsBranchName(message, branchName string) bool {
	msgLower := strings.ToLower(message)
	branchLower := strings.ToLower(branchName)

	// Common merge message patterns
	patterns := []string{
		fmt.Sprintf("merge branch '%s'", branchLower),
		fmt.Sprintf("merge branch \"%s\"", branchLower),
		fmt.Sprintf("merge %s", branchLower),
		fmt.Sprintf("merged %s", branchLower),
		fmt.Sprintf("from %s", branchLower),
	}

	for _, pattern := range patterns {
		if strings.Contains(msgLower, pattern) {
			return true
		}
	}

	return false
}
