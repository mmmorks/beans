package cmd

import (
	"testing"

	"hmans.dev/beans/internal/bean"
)

func TestParseLink(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantType   string
		wantTarget string
		wantErr    bool
	}{
		{
			name:       "valid blocks link",
			input:      "blocks:abc123",
			wantType:   "blocks",
			wantTarget: "abc123",
			wantErr:    false,
		},
		{
			name:       "valid parent link",
			input:      "parent:epic-1",
			wantType:   "parent",
			wantTarget: "epic-1",
			wantErr:    false,
		},
		{
			name:       "valid related link",
			input:      "related:other-bean",
			wantType:   "related",
			wantTarget: "other-bean",
			wantErr:    false,
		},
		{
			name:       "valid duplicates link",
			input:      "duplicates:dup-id",
			wantType:   "duplicates",
			wantTarget: "dup-id",
			wantErr:    false,
		},
		{
			name:    "missing colon",
			input:   "blocksabc123",
			wantErr: true,
		},
		{
			name:    "empty type",
			input:   ":abc123",
			wantErr: true,
		},
		{
			name:    "empty target",
			input:   "blocks:",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:       "target with colons",
			input:      "blocks:id:with:colons",
			wantType:   "blocks",
			wantTarget: "id:with:colons",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkType, targetID, err := parseLink(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseLink(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseLink(%q) unexpected error: %v", tt.input, err)
				return
			}

			if linkType != tt.wantType {
				t.Errorf("parseLink(%q) type = %q, want %q", tt.input, linkType, tt.wantType)
			}

			if targetID != tt.wantTarget {
				t.Errorf("parseLink(%q) target = %q, want %q", tt.input, targetID, tt.wantTarget)
			}
		})
	}
}

func TestIsKnownLinkType(t *testing.T) {
	tests := []struct {
		linkType string
		want     bool
	}{
		{"blocks", true},
		{"duplicates", true},
		{"parent", true},
		{"related", true},
		{"unknown", false},
		{"BLOCKS", false}, // case sensitive
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.linkType, func(t *testing.T) {
			got := isKnownLinkType(tt.linkType)
			if got != tt.want {
				t.Errorf("isKnownLinkType(%q) = %v, want %v", tt.linkType, got, tt.want)
			}
		})
	}
}

func TestApplyTags(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		toAdd    []string
		wantTags []string
		wantErr  bool
	}{
		{
			name:     "add single tag",
			initial:  nil,
			toAdd:    []string{"bug"},
			wantTags: []string{"bug"},
		},
		{
			name:     "add multiple tags",
			initial:  nil,
			toAdd:    []string{"bug", "urgent"},
			wantTags: []string{"bug", "urgent"},
		},
		{
			name:     "add to existing tags",
			initial:  []string{"existing"},
			toAdd:    []string{"new"},
			wantTags: []string{"existing", "new"},
		},
		{
			name:     "empty tags list",
			initial:  []string{"existing"},
			toAdd:    []string{},
			wantTags: []string{"existing"},
		},
		{
			name:    "invalid tag with spaces",
			initial: nil,
			toAdd:   []string{"invalid tag"},
			wantErr: true,
		},
		{
			name:     "uppercase tag gets normalized",
			initial:  nil,
			toAdd:    []string{"InvalidTag"},
			wantTags: []string{"invalidtag"}, // normalized to lowercase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bean.Bean{Tags: tt.initial}
			err := applyTags(b, tt.toAdd)

			if tt.wantErr {
				if err == nil {
					t.Errorf("applyTags() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("applyTags() unexpected error: %v", err)
				return
			}

			if len(b.Tags) != len(tt.wantTags) {
				t.Errorf("applyTags() tags count = %d, want %d", len(b.Tags), len(tt.wantTags))
				return
			}

			for i, want := range tt.wantTags {
				if b.Tags[i] != want {
					t.Errorf("applyTags() tags[%d] = %q, want %q", i, b.Tags[i], want)
				}
			}
		})
	}
}

func TestApplyLinks_SelfReference(t *testing.T) {
	b := &bean.Bean{ID: "abc123"}

	_, err := applyLinks(b, []string{"blocks:abc123"})
	if err == nil {
		t.Error("applyLinks() expected error for self-reference, got nil")
		return
	}

	expected := "bean cannot link to itself"
	if err.Error() != expected {
		t.Errorf("applyLinks() error = %q, want %q", err.Error(), expected)
	}
}

func TestFormatCycle(t *testing.T) {
	tests := []struct {
		path []string
		want string
	}{
		{[]string{"a", "b", "c", "a"}, "a → b → c → a"},
		{[]string{"x", "y"}, "x → y"},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		got := formatCycle(tt.path)
		if got != tt.want {
			t.Errorf("formatCycle(%v) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestRemoveLinks(t *testing.T) {
	tests := []struct {
		name      string
		initial   bean.Links
		toRemove  []string
		wantLinks bean.Links
		wantErr   bool
	}{
		{
			name:      "remove single link",
			initial:   bean.Links{{Type: "blocks", Target: "a1"}, {Type: "parent", Target: "epic1"}},
			toRemove:  []string{"blocks:a1"},
			wantLinks: bean.Links{{Type: "parent", Target: "epic1"}},
		},
		{
			name:      "remove multiple links",
			initial:   bean.Links{{Type: "blocks", Target: "a1"}, {Type: "parent", Target: "epic1"}},
			toRemove:  []string{"blocks:a1", "parent:epic1"},
			wantLinks: bean.Links{},
		},
		{
			name:      "remove non-existent link",
			initial:   bean.Links{{Type: "blocks", Target: "a1"}},
			toRemove:  []string{"blocks:nonexistent"},
			wantLinks: bean.Links{{Type: "blocks", Target: "a1"}},
		},
		{
			name:      "empty remove list",
			initial:   bean.Links{{Type: "blocks", Target: "a1"}},
			toRemove:  []string{},
			wantLinks: bean.Links{{Type: "blocks", Target: "a1"}},
		},
		{
			name:     "invalid link format",
			initial:  bean.Links{{Type: "blocks", Target: "a1"}},
			toRemove: []string{"invalid"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bean.Bean{Links: tt.initial}
			err := removeLinks(b, tt.toRemove)

			if tt.wantErr {
				if err == nil {
					t.Errorf("removeLinks() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("removeLinks() unexpected error: %v", err)
				return
			}

			if len(b.Links) != len(tt.wantLinks) {
				t.Errorf("removeLinks() links count = %d, want %d", len(b.Links), len(tt.wantLinks))
				return
			}

			for i, want := range tt.wantLinks {
				if b.Links[i].Type != want.Type || b.Links[i].Target != want.Target {
					t.Errorf("removeLinks() links[%d] = %v, want %v", i, b.Links[i], want)
				}
			}
		})
	}
}
