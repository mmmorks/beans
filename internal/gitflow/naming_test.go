package gitflow

import (
	"strings"
	"testing"
)

func TestBuildBranchName(t *testing.T) {
	tests := []struct {
		name     string
		beanID   string
		slug     string
		expected string
	}{
		{
			name:     "with slug",
			beanID:   "beans-abc123",
			slug:     "user-authentication",
			expected: "beans-abc123/user-authentication",
		},
		{
			name:     "without slug",
			beanID:   "beans-abc123",
			slug:     "",
			expected: "beans-abc123",
		},
		{
			name:     "slug needs sanitization",
			beanID:   "beans-xyz",
			slug:     "Fix Bug #123!",
			expected: "beans-xyz/fix-bug-123",
		},
		{
			name:     "slug with special characters",
			beanID:   "beans-test",
			slug:     "Add @mention & <tags>",
			expected: "beans-test/add-mention-tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildBranchName(tt.beanID, tt.slug)
			if result != tt.expected {
				t.Errorf("BuildBranchName(%q, %q) = %q, want %q", tt.beanID, tt.slug, result, tt.expected)
			}
		})
	}
}

func TestParseBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		wantID     string
		wantOk     bool
	}{
		{
			name:       "full format with slug",
			branchName: "beans-abc123/user-authentication",
			wantID:     "beans-abc123",
			wantOk:     true,
		},
		{
			name:       "bean ID only",
			branchName: "beans-abc123",
			wantID:     "beans-abc123",
			wantOk:     true,
		},
		{
			name:       "without prefix",
			branchName: "abc123/feature",
			wantID:     "abc123",
			wantOk:     true,
		},
		{
			name:       "extracts first word as ID",
			branchName: "feature/test",
			wantID:     "feature",
			wantOk:     true,
		},
		{
			name:       "invalid - empty string",
			branchName: "",
			wantID:     "",
			wantOk:     false,
		},
		{
			name:       "different prefix",
			branchName: "task-xyz789/implement-feature",
			wantID:     "task-xyz789",
			wantOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOk := ParseBranchName(tt.branchName)
			if gotOk != tt.wantOk {
				t.Errorf("ParseBranchName(%q) ok = %v, want %v", tt.branchName, gotOk, tt.wantOk)
			}
			if gotID != tt.wantID {
				t.Errorf("ParseBranchName(%q) id = %q, want %q", tt.branchName, gotID, tt.wantID)
			}
		})
	}
}

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already clean",
			input:    "user-authentication",
			expected: "user-authentication",
		},
		{
			name:     "mixed case to lowercase",
			input:    "UserAuthentication",
			expected: "userauthentication",
		},
		{
			name:     "spaces to hyphens",
			input:    "user authentication system",
			expected: "user-authentication-system",
		},
		{
			name:     "special characters removed",
			input:    "fix: bug #123!",
			expected: "fix-bug-123",
		},
		{
			name:     "consecutive hyphens collapsed",
			input:    "foo---bar--baz",
			expected: "foo-bar-baz",
		},
		{
			name:     "trim leading/trailing hyphens",
			input:    "-test-slug-",
			expected: "test-slug",
		},
		{
			name:     "unicode characters removed",
			input:    "café-résumé",
			expected: "caf-r-sum",
		},
		{
			name:     "multiple special chars",
			input:    "Add @mentions & <tags> [feature]",
			expected: "add-mentions-tags-feature",
		},
		{
			name:     "very long slug gets truncated",
			input:    "this-is-a-very-long-slug-that-should-be-truncated-to-fifty-characters-maximum",
			expected: "this-is-a-very-long-slug-that-should-be-truncated",
		},
		{
			name:     "truncation removes trailing hyphen",
			input:    "this-is-a-very-long-slug-that-ends-with-hyphen-at-exactly-fifty-chars-position",
			expected: "this-is-a-very-long-slug-that-ends-with-hyphen-at",
		},
		{
			name:     "empty after sanitization",
			input:    "@#$%",
			expected: "",
		},
		{
			name:     "underscores converted to hyphens",
			input:    "test_feature_name",
			expected: "test-feature-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeSlug(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// Verify length constraint
			if len(result) > 50 {
				t.Errorf("SanitizeSlug(%q) length = %d, want <= 50", tt.input, len(result))
			}
			// Verify no leading/trailing hyphens
			if len(result) > 0 && (result[0] == '-' || result[len(result)-1] == '-') {
				t.Errorf("SanitizeSlug(%q) has leading/trailing hyphen: %q", tt.input, result)
			}
		})
	}
}

func TestContainsBranchName(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		branchName string
		want       bool
	}{
		{
			name:       "merge branch with single quotes",
			message:    "Merge branch 'beans-abc123/feature'",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "merge branch with double quotes",
			message:    "Merge branch \"beans-abc123/feature\"",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "merge without quotes",
			message:    "Merge beans-abc123/feature into main",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "merged past tense",
			message:    "Merged beans-abc123/feature via PR #123",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "from pattern",
			message:    "Merge pull request #123 from beans-abc123/feature",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "case insensitive match",
			message:    "MERGE BRANCH 'BEANS-ABC123/FEATURE'",
			branchName: "beans-abc123/feature",
			want:       true,
		},
		{
			name:       "no match - different branch",
			message:    "Merge branch 'other-branch'",
			branchName: "beans-abc123/feature",
			want:       false,
		},
		{
			name:       "no match - not a merge message",
			message:    "feat: add new feature",
			branchName: "beans-abc123/feature",
			want:       false,
		},
		{
			name:       "partial match not accepted",
			message:    "Merge branch 'beans-abc123/feature-extended'",
			branchName: "beans-abc123/feature",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsBranchName(tt.message, tt.branchName)
			if result != tt.want {
				t.Errorf("containsBranchName(%q, %q) = %v, want %v", tt.message, tt.branchName, result, tt.want)
			}
		})
	}
}

func TestSanitizeSlugLength(t *testing.T) {
	// Test that exactly 50 chars is preserved
	input := strings.Repeat("a", 50)
	result := SanitizeSlug(input)
	if len(result) != 50 {
		t.Errorf("SanitizeSlug with 50 chars: got len %d, want 50", len(result))
	}

	// Test that 51 chars gets truncated
	input = strings.Repeat("a", 51)
	result = SanitizeSlug(input)
	if len(result) != 50 {
		t.Errorf("SanitizeSlug with 51 chars: got len %d, want 50", len(result))
	}
}

func TestSanitizeSlugConsecutiveHyphens(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a--b", "a-b"},
		{"a---b", "a-b"},
		{"a----b", "a-b"},
		{"a-b-c", "a-b-c"},
		{"--a--b--", "a-b"},
	}

	for _, tt := range tests {
		result := SanitizeSlug(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeSlug(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
