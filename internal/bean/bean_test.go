package bean

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedTitle  string
		expectedStatus string
		expectedBody   string
		wantErr        bool
	}{
		{
			name: "basic bean",
			input: `---
title: Test Bean
status: todo
---

This is the body.`,
			expectedTitle:  "Test Bean",
			expectedStatus: "todo",
			expectedBody:   "\nThis is the body.",
		},
		{
			name: "with timestamps",
			input: `---
title: With Times
status: in-progress
created_at: 2024-01-15T10:30:00Z
updated_at: 2024-01-16T14:45:00Z
---

Body content here.`,
			expectedTitle:  "With Times",
			expectedStatus: "in-progress",
			expectedBody:   "\nBody content here.",
		},
		{
			name: "empty body",
			input: `---
title: No Body
status: completed
---`,
			expectedTitle:  "No Body",
			expectedStatus: "completed",
			expectedBody:   "",
		},
		{
			name: "multiline body",
			input: `---
title: Multi Line
status: todo
---

# Header

- Item 1
- Item 2

Paragraph text.`,
			expectedTitle:  "Multi Line",
			expectedStatus: "todo",
			expectedBody:   "\n# Header\n\n- Item 1\n- Item 2\n\nParagraph text.",
		},
		{
			name:           "plain text without frontmatter",
			input:          `Just plain text without any YAML frontmatter.`,
			expectedTitle:  "",
			expectedStatus: "",
			expectedBody:   "Just plain text without any YAML frontmatter.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bean, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bean.Title != tt.expectedTitle {
				t.Errorf("Title = %q, want %q", bean.Title, tt.expectedTitle)
			}
			if bean.Status != tt.expectedStatus {
				t.Errorf("Status = %q, want %q", bean.Status, tt.expectedStatus)
			}
			if bean.Body != tt.expectedBody {
				t.Errorf("Body = %q, want %q", bean.Body, tt.expectedBody)
			}
		})
	}
}

func TestParseWithType(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
	}{
		{
			name: "with type field",
			input: `---
title: Bug Report
status: todo
type: bug
---

Description of the bug.`,
			expectedType: "bug",
		},
		{
			name: "without type field",
			input: `---
title: No Type
status: todo
---

No type specified.`,
			expectedType: "",
		},
		{
			// Backwards compatibility: beans with types not in current config
			// should still be readable without error
			name: "with unknown type (backwards compatibility)",
			input: `---
title: Legacy Bean
status: todo
type: deprecated-type-no-longer-in-config
---`,
			expectedType: "deprecated-type-no-longer-in-config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bean, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bean.Type != tt.expectedType {
				t.Errorf("Type = %q, want %q", bean.Type, tt.expectedType)
			}
		})
	}
}

func TestParseWithPriority(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedPriority string
	}{
		{
			name: "with priority field",
			input: `---
title: Urgent Bug
status: todo
type: bug
priority: critical
---

Fix this immediately.`,
			expectedPriority: "critical",
		},
		{
			name: "without priority field",
			input: `---
title: Normal Task
status: todo
---

No priority specified.`,
			expectedPriority: "",
		},
		{
			name: "with high priority",
			input: `---
title: Important Feature
status: in-progress
priority: high
---`,
			expectedPriority: "high",
		},
		{
			name: "with deferred priority",
			input: `---
title: Later Task
status: backlog
priority: deferred
---`,
			expectedPriority: "deferred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bean, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bean.Priority != tt.expectedPriority {
				t.Errorf("Priority = %q, want %q", bean.Priority, tt.expectedPriority)
			}
		})
	}
}

func TestRenderWithPriority(t *testing.T) {
	tests := []struct {
		name     string
		bean     *Bean
		contains []string
	}{
		{
			name: "with priority",
			bean: &Bean{
				Title:    "High Priority",
				Status:   "todo",
				Priority: "high",
			},
			contains: []string{
				"title: High Priority",
				"status: todo",
				"priority: high",
			},
		},
		{
			name: "without priority",
			bean: &Bean{
				Title:  "No Priority",
				Status: "todo",
			},
			contains: []string{
				"title: No Priority",
				"status: todo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := tt.bean.Render()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content := string(rendered)
			for _, want := range tt.contains {
				if !strings.Contains(content, want) {
					t.Errorf("Render() missing %q in:\n%s", want, content)
				}
			}

			// Verify priority is NOT in output when empty
			if tt.bean.Priority == "" && strings.Contains(content, "priority:") {
				t.Errorf("Render() should not contain 'priority:' when priority is empty:\n%s", content)
			}
		})
	}
}

func TestPriorityRoundtrip(t *testing.T) {
	priorities := []string{"critical", "high", "normal", "low", "deferred", ""}

	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			original := &Bean{
				Title:    "Test Bean",
				Status:   "todo",
				Priority: priority,
			}

			rendered, err := original.Render()
			if err != nil {
				t.Fatalf("Render() error: %v", err)
			}

			parsed, err := Parse(strings.NewReader(string(rendered)))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			if parsed.Priority != original.Priority {
				t.Errorf("Priority roundtrip failed: got %q, want %q", parsed.Priority, original.Priority)
			}
		})
	}
}

func TestRender(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		bean     *Bean
		contains []string
	}{
		{
			name: "basic bean",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
			},
			contains: []string{
				"---",
				"title: Test Bean",
				"status: todo",
			},
		},
		{
			name: "with body",
			bean: &Bean{
				Title:  "With Body",
				Status: "completed",
				Body:   "This is content.",
			},
			contains: []string{
				"title: With Body",
				"status: completed",
				"This is content.",
			},
		},
		{
			name: "with timestamps",
			bean: &Bean{
				Title:     "Timed",
				Status:    "todo",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
			contains: []string{
				"title: Timed",
				"created_at:",
				"updated_at:",
			},
		},
		{
			name: "with type",
			bean: &Bean{
				Title:  "Typed Bean",
				Status: "todo",
				Type:   "bug",
			},
			contains: []string{
				"title: Typed Bean",
				"status: todo",
				"type: bug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.bean.Render()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, result)
				}
			}
		})
	}
}

func TestParseRenderRoundtrip(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	later := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)

	tests := []struct {
		name string
		bean *Bean
	}{
		{
			name: "basic",
			bean: &Bean{
				Title:  "Basic Bean",
				Status: "todo",
			},
		},
		{
			name: "with body",
			bean: &Bean{
				Title:  "Bean With Body",
				Status: "in-progress",
				Body:   "This is the body content.\n\nWith multiple paragraphs.",
			},
		},
		{
			name: "with timestamps",
			bean: &Bean{
				Title:     "Timestamped Bean",
				Status:    "completed",
				CreatedAt: &now,
				UpdatedAt: &later,
				Body:      "Some content.",
			},
		},
		{
			name: "with type",
			bean: &Bean{
				Title:  "Typed Bean",
				Status: "todo",
				Type:   "bug",
				Body:   "Bug description.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Render to bytes
			rendered, err := tt.bean.Render()
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}

			// Parse back
			parsed, err := Parse(strings.NewReader(string(rendered)))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Compare fields
			if parsed.Title != tt.bean.Title {
				t.Errorf("Title roundtrip: got %q, want %q", parsed.Title, tt.bean.Title)
			}
			if parsed.Status != tt.bean.Status {
				t.Errorf("Status roundtrip: got %q, want %q", parsed.Status, tt.bean.Status)
			}
			if parsed.Type != tt.bean.Type {
				t.Errorf("Type roundtrip: got %q, want %q", parsed.Type, tt.bean.Type)
			}

			// Body comparison (parse adds newline prefix for non-empty body)
			wantBody := tt.bean.Body
			if wantBody != "" {
				wantBody = "\n" + wantBody
			}
			if parsed.Body != wantBody {
				t.Errorf("Body roundtrip: got %q, want %q", parsed.Body, wantBody)
			}

			// Timestamp comparison
			if tt.bean.CreatedAt != nil {
				if parsed.CreatedAt == nil {
					t.Error("CreatedAt: got nil, want non-nil")
				} else if !parsed.CreatedAt.Equal(*tt.bean.CreatedAt) {
					t.Errorf("CreatedAt: got %v, want %v", parsed.CreatedAt, tt.bean.CreatedAt)
				}
			}
			if tt.bean.UpdatedAt != nil {
				if parsed.UpdatedAt == nil {
					t.Error("UpdatedAt: got nil, want non-nil")
				} else if !parsed.UpdatedAt.Equal(*tt.bean.UpdatedAt) {
					t.Errorf("UpdatedAt: got %v, want %v", parsed.UpdatedAt, tt.bean.UpdatedAt)
				}
			}
		})
	}
}

func TestBeanJSONSerialization(t *testing.T) {
	t.Run("body omitted when empty", func(t *testing.T) {
		bean := &Bean{
			ID:     "test-123",
			Title:  "Test Bean",
			Status: "todo",
			Body:   "",
		}

		data, err := json.Marshal(bean)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		jsonStr := string(data)
		if strings.Contains(jsonStr, `"body"`) {
			t.Errorf("JSON should not contain 'body' field when empty, got: %s", jsonStr)
		}
	})

	t.Run("body included when non-empty", func(t *testing.T) {
		bean := &Bean{
			ID:     "test-123",
			Title:  "Test Bean",
			Status: "todo",
			Body:   "This is the body content.",
		}

		data, err := json.Marshal(bean)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"body":"This is the body content."`) {
			t.Errorf("JSON should contain 'body' field with content, got: %s", jsonStr)
		}
	})
}

func TestParseWithLinks(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLinks Links
	}{
		{
			name: "single link",
			input: `---
title: Test
status: todo
links:
  - blocks: abc123
---`,
			expectedLinks: Links{
				{Type: "blocks", Target: "abc123"},
			},
		},
		{
			name: "multiple links of same type",
			input: `---
title: Test
status: todo
links:
  - blocks: abc123
  - blocks: def456
---`,
			expectedLinks: Links{
				{Type: "blocks", Target: "abc123"},
				{Type: "blocks", Target: "def456"},
			},
		},
		{
			name: "multiple link types",
			input: `---
title: Test
status: todo
links:
  - blocks: abc123
  - parent: xyz789
---`,
			expectedLinks: Links{
				{Type: "blocks", Target: "abc123"},
				{Type: "parent", Target: "xyz789"},
			},
		},
		{
			name: "no links",
			input: `---
title: Test
status: todo
---`,
			expectedLinks: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bean, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tt.expectedLinks) == 0 && len(bean.Links) == 0 {
				return // Both empty, OK
			}

			if len(bean.Links) != len(tt.expectedLinks) {
				t.Errorf("Links count = %d, want %d", len(bean.Links), len(tt.expectedLinks))
				return
			}

			for i, expected := range tt.expectedLinks {
				actual := bean.Links[i]
				if actual.Type != expected.Type || actual.Target != expected.Target {
					t.Errorf("Links[%d] = {%s, %s}, want {%s, %s}",
						i, actual.Type, actual.Target, expected.Type, expected.Target)
				}
			}
		})
	}
}

func TestRenderWithLinks(t *testing.T) {
	tests := []struct {
		name     string
		bean     *Bean
		contains []string
	}{
		{
			name: "with single link",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
				Links: Links{
					{Type: "blocks", Target: "abc123"},
				},
			},
			contains: []string{
				"links:",
				"- blocks: abc123",
			},
		},
		{
			name: "with multiple links",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
				Links: Links{
					{Type: "blocks", Target: "abc123"},
					{Type: "blocks", Target: "def456"},
					{Type: "parent", Target: "xyz789"},
				},
			},
			contains: []string{
				"links:",
				"- blocks: abc123",
				"- blocks: def456",
				"- parent: xyz789",
			},
		},
		{
			name: "without links",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
			},
			contains: []string{
				"title: Test Bean",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.bean.Render()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, result)
				}
			}

			// Check that empty links don't appear in output
			if tt.bean.Links == nil && strings.Contains(result, "links:") {
				t.Errorf("output should not contain 'links:' when no links\ngot:\n%s", result)
			}
		})
	}
}

func TestLinksRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		links Links
	}{
		{
			name: "single link",
			links: Links{
				{Type: "blocks", Target: "abc123"},
			},
		},
		{
			name: "multiple links same type",
			links: Links{
				{Type: "blocks", Target: "abc123"},
				{Type: "blocks", Target: "def456"},
			},
		},
		{
			name: "multiple link types",
			links: Links{
				{Type: "blocks", Target: "abc123"},
				{Type: "parent", Target: "xyz789"},
				{Type: "related", Target: "foo"},
				{Type: "related", Target: "bar"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &Bean{
				Title:  "Test",
				Status: "todo",
				Links:  tt.links,
			}

			rendered, err := original.Render()
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}

			parsed, err := Parse(strings.NewReader(string(rendered)))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(parsed.Links) != len(tt.links) {
				t.Errorf("Links count: got %d, want %d", len(parsed.Links), len(tt.links))
				return
			}

			for i, expected := range tt.links {
				actual := parsed.Links[i]
				if actual.Type != expected.Type || actual.Target != expected.Target {
					t.Errorf("Links[%d] = {%s, %s}, want {%s, %s}",
						i, actual.Type, actual.Target, expected.Type, expected.Target)
				}
			}
		})
	}
}

func TestLinksHelperMethods(t *testing.T) {
	links := Links{
		{Type: "blocks", Target: "abc"},
		{Type: "blocks", Target: "def"},
		{Type: "parent", Target: "xyz"},
	}

	t.Run("HasType", func(t *testing.T) {
		if !links.HasType("blocks") {
			t.Error("expected HasType('blocks') = true")
		}
		if !links.HasType("parent") {
			t.Error("expected HasType('parent') = true")
		}
		if links.HasType("nonexistent") {
			t.Error("expected HasType('nonexistent') = false")
		}
	})

	t.Run("HasLink", func(t *testing.T) {
		if !links.HasLink("blocks", "abc") {
			t.Error("expected HasLink('blocks', 'abc') = true")
		}
		if !links.HasLink("parent", "xyz") {
			t.Error("expected HasLink('parent', 'xyz') = true")
		}
		if links.HasLink("blocks", "xyz") {
			t.Error("expected HasLink('blocks', 'xyz') = false")
		}
		if links.HasLink("parent", "abc") {
			t.Error("expected HasLink('parent', 'abc') = false")
		}
	})

	t.Run("Targets", func(t *testing.T) {
		targets := links.Targets("blocks")
		if len(targets) != 2 || targets[0] != "abc" || targets[1] != "def" {
			t.Errorf("Targets('blocks') = %v, want [abc def]", targets)
		}
		targets = links.Targets("parent")
		if len(targets) != 1 || targets[0] != "xyz" {
			t.Errorf("Targets('parent') = %v, want [xyz]", targets)
		}
		targets = links.Targets("nonexistent")
		if len(targets) != 0 {
			t.Errorf("Targets('nonexistent') = %v, want []", targets)
		}
	})

	t.Run("Add", func(t *testing.T) {
		newLinks := links.Add("blocks", "ghi")
		if len(newLinks) != 4 {
			t.Errorf("Add new link: got len=%d, want 4", len(newLinks))
		}
		// Adding duplicate should not add
		sameLinks := links.Add("blocks", "abc")
		if len(sameLinks) != 3 {
			t.Errorf("Add duplicate: got len=%d, want 3", len(sameLinks))
		}
	})

	t.Run("Remove", func(t *testing.T) {
		newLinks := links.Remove("blocks", "abc")
		if len(newLinks) != 2 {
			t.Errorf("Remove existing: got len=%d, want 2", len(newLinks))
		}
		if newLinks.HasLink("blocks", "abc") {
			t.Error("Remove didn't remove the link")
		}
		// Removing non-existent should not change anything
		sameLinks := links.Remove("blocks", "nonexistent")
		if len(sameLinks) != 3 {
			t.Errorf("Remove non-existent: got len=%d, want 3", len(sameLinks))
		}
	})
}

func TestValidateTag(t *testing.T) {
	tests := []struct {
		tag     string
		wantErr bool
	}{
		{"frontend", false},
		{"backend", false},
		{"tech-debt", false},
		{"v1", false},
		{"a", false},
		{"urgent2", false},
		{"wont-fix", false},
		{"a-b-c", false},
		{"", true},           // empty
		{"Frontend", true},   // uppercase
		{"URGENT", true},     // all uppercase
		{"123", true},        // starts with number
		{"123abc", true},     // starts with number
		{"my tag", true},     // contains space
		{"my_tag", true},     // contains underscore
		{"my--tag", true},    // consecutive hyphens
		{"-tag", true},       // starts with hyphen
		{"tag-", true},       // ends with hyphen
		{"my.tag", true},     // contains dot
		{"my/tag", true},     // contains slash
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			err := ValidateTag(tt.tag)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateTag(%q) = nil, want error", tt.tag)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateTag(%q) = %v, want nil", tt.tag, err)
			}
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"frontend", "frontend"},
		{"FRONTEND", "frontend"},
		{"FrontEnd", "frontend"},
		{"  frontend  ", "frontend"},
		{"  FRONTEND  ", "frontend"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeTag(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeTag(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBeanTagMethods(t *testing.T) {
	t.Run("HasTag", func(t *testing.T) {
		b := &Bean{Tags: []string{"frontend", "urgent"}}
		if !b.HasTag("frontend") {
			t.Error("expected HasTag('frontend') = true")
		}
		if !b.HasTag("urgent") {
			t.Error("expected HasTag('urgent') = true")
		}
		if b.HasTag("backend") {
			t.Error("expected HasTag('backend') = false")
		}
		// Case insensitive lookup
		if !b.HasTag("FRONTEND") {
			t.Error("expected HasTag('FRONTEND') = true (case insensitive)")
		}
	})

	t.Run("AddTag", func(t *testing.T) {
		b := &Bean{Tags: []string{"frontend"}}

		// Add new valid tag
		if err := b.AddTag("backend"); err != nil {
			t.Errorf("AddTag('backend') error: %v", err)
		}
		if len(b.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(b.Tags))
		}

		// Adding duplicate should not add
		if err := b.AddTag("frontend"); err != nil {
			t.Errorf("AddTag('frontend') error: %v", err)
		}
		if len(b.Tags) != 2 {
			t.Errorf("expected 2 tags (no duplicate), got %d", len(b.Tags))
		}

		// Adding invalid tag should error
		if err := b.AddTag("Invalid Tag"); err == nil {
			t.Error("expected AddTag('Invalid Tag') to error")
		}
	})

	t.Run("RemoveTag", func(t *testing.T) {
		b := &Bean{Tags: []string{"frontend", "backend", "urgent"}}

		b.RemoveTag("backend")
		if len(b.Tags) != 2 {
			t.Errorf("expected 2 tags after remove, got %d", len(b.Tags))
		}
		if b.HasTag("backend") {
			t.Error("expected backend tag to be removed")
		}

		// Case insensitive removal
		b.RemoveTag("FRONTEND")
		if len(b.Tags) != 1 {
			t.Errorf("expected 1 tag after remove, got %d", len(b.Tags))
		}
		if b.HasTag("frontend") {
			t.Error("expected frontend tag to be removed")
		}

		// Remove non-existent tag (should not error)
		b.RemoveTag("nonexistent")
		if len(b.Tags) != 1 {
			t.Errorf("expected 1 tag (no change), got %d", len(b.Tags))
		}
	})
}

func TestParseWithTags(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedTags []string
	}{
		{
			name: "single tag",
			input: `---
title: Test
status: todo
tags:
  - frontend
---`,
			expectedTags: []string{"frontend"},
		},
		{
			name: "multiple tags",
			input: `---
title: Test
status: todo
tags:
  - frontend
  - urgent
  - tech-debt
---`,
			expectedTags: []string{"frontend", "urgent", "tech-debt"},
		},
		{
			name: "inline tags syntax",
			input: `---
title: Test
status: todo
tags: [frontend, backend]
---`,
			expectedTags: []string{"frontend", "backend"},
		},
		{
			name: "no tags",
			input: `---
title: Test
status: todo
---`,
			expectedTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bean, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tt.expectedTags) == 0 && len(bean.Tags) == 0 {
				return // Both empty, OK
			}

			if len(bean.Tags) != len(tt.expectedTags) {
				t.Errorf("Tags count = %d, want %d", len(bean.Tags), len(tt.expectedTags))
				return
			}

			for i, expected := range tt.expectedTags {
				if bean.Tags[i] != expected {
					t.Errorf("Tags[%d] = %q, want %q", i, bean.Tags[i], expected)
				}
			}
		})
	}
}

func TestRenderWithTags(t *testing.T) {
	tests := []struct {
		name     string
		bean     *Bean
		contains []string
	}{
		{
			name: "with single tag",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
				Tags:   []string{"frontend"},
			},
			contains: []string{
				"tags:",
				"- frontend",
			},
		},
		{
			name: "with multiple tags",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
				Tags:   []string{"frontend", "urgent", "tech-debt"},
			},
			contains: []string{
				"tags:",
				"- frontend",
				"- urgent",
				"- tech-debt",
			},
		},
		{
			name: "without tags",
			bean: &Bean{
				Title:  "Test Bean",
				Status: "todo",
			},
			contains: []string{
				"title: Test Bean",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.bean.Render()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, result)
				}
			}

			// Check that empty tags don't appear in output
			if len(tt.bean.Tags) == 0 && strings.Contains(result, "tags:") {
				t.Errorf("output should not contain 'tags:' when no tags\ngot:\n%s", result)
			}
		})
	}
}

func TestTagsRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		tags []string
	}{
		{
			name: "single tag",
			tags: []string{"frontend"},
		},
		{
			name: "multiple tags",
			tags: []string{"frontend", "backend", "urgent"},
		},
		{
			name: "hyphenated tags",
			tags: []string{"tech-debt", "wont-fix", "needs-review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &Bean{
				Title:  "Test",
				Status: "todo",
				Tags:   tt.tags,
			}

			rendered, err := original.Render()
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}

			parsed, err := Parse(strings.NewReader(string(rendered)))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if len(parsed.Tags) != len(tt.tags) {
				t.Errorf("Tags count: got %d, want %d", len(parsed.Tags), len(tt.tags))
				return
			}

			for i, expected := range tt.tags {
				if parsed.Tags[i] != expected {
					t.Errorf("Tags[%d] = %q, want %q", i, parsed.Tags[i], expected)
				}
			}
		})
	}
}
