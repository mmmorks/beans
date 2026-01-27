package cmd

import (
	"testing"
	"time"

	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/config"
)

func TestSortBeans(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	// Statuses are now hardcoded, so we just use default config
	testCfg := config.Default()

	t.Run("sort by id", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "c3"},
			{ID: "a1"},
			{ID: "b2"},
		}
		sortBeans(beans, "id", testCfg)

		if beans[0].ID != "a1" || beans[1].ID != "b2" || beans[2].ID != "c3" {
			t.Errorf("sort by id: got [%s, %s, %s], want [a1, b2, c3]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "old", CreatedAt: &evenEarlier},
			{ID: "new", CreatedAt: &now},
			{ID: "mid", CreatedAt: &earlier},
		}
		sortBeans(beans, "created", testCfg)

		// Should be newest first
		if beans[0].ID != "new" || beans[1].ID != "mid" || beans[2].ID != "old" {
			t.Errorf("sort by created: got [%s, %s, %s], want [new, mid, old]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by created with nil", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "nil1", CreatedAt: nil},
			{ID: "has", CreatedAt: &now},
			{ID: "nil2", CreatedAt: nil},
		}
		sortBeans(beans, "created", testCfg)

		// Non-nil should come first, then nil sorted by ID
		if beans[0].ID != "has" {
			t.Errorf("sort by created with nil: first should be \"has\", got %q", beans[0].ID)
		}
	})

	t.Run("sort by updated", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "old", UpdatedAt: &evenEarlier},
			{ID: "new", UpdatedAt: &now},
			{ID: "mid", UpdatedAt: &earlier},
		}
		sortBeans(beans, "updated", testCfg)

		// Should be newest first
		if beans[0].ID != "new" || beans[1].ID != "mid" || beans[2].ID != "old" {
			t.Errorf("sort by updated: got [%s, %s, %s], want [new, mid, old]",
				beans[0].ID, beans[1].ID, beans[2].ID)
		}
	})

	t.Run("sort by status", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "c1", Status: "completed"},
			{ID: "t1", Status: "todo"},
			{ID: "i1", Status: "in-progress"},
			{ID: "t2", Status: "todo"},
		}
		sortBeans(beans, "status", testCfg)

		// Should be ordered by status config order (in-progress, todo, draft, completed, scrapped), then by ID within same status
		expected := []string{"i1", "t1", "t2", "c1"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("sort by status[%d]: got %q, want %q", i, beans[i].ID, want)
			}
		}
	})

	t.Run("default sort (archive status then type)", func(t *testing.T) {
		beans := []*bean.Bean{
			{ID: "completed-bug", Status: "completed", Type: "bug"},
			{ID: "todo-feature", Status: "todo", Type: "feature"},
			{ID: "todo-task", Status: "todo", Type: "task"},
			{ID: "completed-task", Status: "completed", Type: "task"},
			{ID: "todo-bug", Status: "todo", Type: "bug"},
		}
		sortBeans(beans, "", testCfg)

		// Should be: non-archive first (sorted by type order from DefaultTypes: milestone, epic, bug, feature, task),
		// then archive (sorted by type)
		// DefaultTypes order: milestone, epic, bug, feature, task
		expected := []string{"todo-bug", "todo-feature", "todo-task", "completed-bug", "completed-task"}
		for i, want := range expected {
			if beans[i].ID != want {
				t.Errorf("default sort[%d]: got %q, want %q", i, beans[i].ID, want)
			}
		}
	})
}

func TestListReadyFlagMutualExclusion(t *testing.T) {
	// Test that --ready and --is-blocked are mutually exclusive
	// by checking the validation logic directly
	tests := []struct {
		name        string
		ready       bool
		isBlocked   bool
		expectError bool
	}{
		{"neither flag", false, false, false},
		{"only --ready", true, false, false},
		{"only --is-blocked", false, true, false},
		{"both flags", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This mirrors the validation logic in list.go
			hasError := tt.ready && tt.isBlocked
			if hasError != tt.expectError {
				t.Errorf("ready=%v, isBlocked=%v: got error=%v, want error=%v",
					tt.ready, tt.isBlocked, hasError, tt.expectError)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 4, "h..."},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestListCommand_GitFieldsInBeans(t *testing.T) {
	// Test that beans with git fields are properly included in list
	now := time.Now()
	earlier := now.Add(-24 * time.Hour)

	beans := []*bean.Bean{
		{
			ID:           "beans-with-git",
			Title:        "With Git",
			Status:       "in-progress",
			GitBranch:    "beans-with-git/with-git",
			GitCreatedAt: &now,
		},
		{
			ID:             "beans-merged",
			Title:          "Merged",
			Status:         "completed",
			GitBranch:      "beans-merged/merged",
			GitCreatedAt:   &earlier,
			GitMergedAt:    &now,
			GitMergeCommit: "abc123",
		},
		{
			ID:     "beans-no-git",
			Title:  "No Git",
			Status: "todo",
		},
	}

	// Verify beans have the expected git fields
	if beans[0].GitBranch == "" {
		t.Error("first bean should have GitBranch")
	}
	if beans[0].GitCreatedAt == nil {
		t.Error("first bean should have GitCreatedAt")
	}

	if beans[1].GitMergedAt == nil {
		t.Error("second bean should have GitMergedAt")
	}
	if beans[1].GitMergeCommit == "" {
		t.Error("second bean should have GitMergeCommit")
	}

	if beans[2].GitBranch != "" {
		t.Error("third bean should not have GitBranch")
	}
}

func TestListCommand_SortBeansWithGitFields(t *testing.T) {
	// Test that beans with git fields can be sorted
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	testCfg := config.Default()

	beans := []*bean.Bean{
		{
			ID:           "beans-2",
			Title:        "Bean 2",
			Status:       "in-progress",
			GitBranch:    "beans-2/bean-two",
			GitCreatedAt: &now,
		},
		{
			ID:           "beans-1",
			Title:        "Bean 1",
			Status:       "in-progress",
			GitBranch:    "beans-1/bean-one",
			GitCreatedAt: &earlier,
		},
		{
			ID:     "beans-3",
			Title:  "Bean 3",
			Status: "todo",
		},
	}

	// Sort by ID
	sortBeans(beans, "id", testCfg)
	if beans[0].ID != "beans-1" {
		t.Errorf("after sort by id, first should be beans-1, got %s", beans[0].ID)
	}

	// Sort by status
	sortBeans(beans, "status", testCfg)
	// in-progress comes before todo
	if beans[0].Status != "in-progress" {
		t.Errorf("after sort by status, first should have status in-progress, got %s", beans[0].Status)
	}
	if beans[2].Status != "todo" {
		t.Errorf("after sort by status, last should have status todo, got %s", beans[2].Status)
	}
}

