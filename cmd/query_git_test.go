package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/graph/model"
)

// setupQueryGitTestCore creates a test core with beans that have git metadata
func setupQueryGitTestCore(t *testing.T) (*beancore.Core, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	beansDir := filepath.Join(tmpDir, ".beans")

	// Create the .beans directory
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatalf("failed to create .beans dir: %v", err)
	}

	cfg := config.Default()
	testCore := beancore.New(beansDir, cfg)
	testCore.SetWarnWriter(nil)
	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	oldCore := core
	core = testCore

	cleanup := func() {
		core = oldCore
	}

	return testCore, cleanup
}

func TestQueryCommand_GitBranchFilter(t *testing.T) {
	testCore, cleanup := setupQueryGitTestCore(t)
	defer cleanup()

	// Create beans with and without git branches
	now := time.Now()

	withBranch := &bean.Bean{
		ID:           "beans-with-branch",
		Slug:         "with-branch",
		Title:        "With Branch",
		Status:       "in-progress",
		GitBranch:    "beans-with-branch/with-branch",
		GitCreatedAt: &now,
	}
	testCore.Create(withBranch)

	withoutBranch := &bean.Bean{
		ID:     "beans-without-branch",
		Slug:   "without-branch",
		Title:  "Without Branch",
		Status: "todo",
	}
	testCore.Create(withoutBranch)

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	qr := resolver.Query()

	t.Run("filter hasGitBranch=true", func(t *testing.T) {
		hasGitBranch := true
		filter := &model.BeanFilter{
			HasGitBranch: &hasGitBranch,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 bean with git branch, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-with-branch" {
			t.Errorf("expected beans-with-branch, got %s", beans[0].ID)
		}
	})

	t.Run("filter hasGitBranch=false", func(t *testing.T) {
		hasGitBranch := false
		filter := &model.BeanFilter{
			HasGitBranch: &hasGitBranch,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 bean without git branch, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-without-branch" {
			t.Errorf("expected beans-without-branch, got %s", beans[0].ID)
		}
	})
}

func TestQueryCommand_GitBranchMergedFilter(t *testing.T) {
	testCore, cleanup := setupQueryGitTestCore(t)
	defer cleanup()

	// Create beans with different git merge states
	now := time.Now()

	merged := &bean.Bean{
		ID:             "beans-merged",
		Slug:           "merged",
		Title:          "Merged",
		Status:         "completed",
		GitBranch:      "beans-merged/merged",
		GitMergedAt:    &now,
		GitMergeCommit: "abc123",
	}
	testCore.Create(merged)

	unmerged := &bean.Bean{
		ID:           "beans-unmerged",
		Slug:         "unmerged",
		Title:        "Unmerged",
		Status:       "in-progress",
		GitBranch:    "beans-unmerged/unmerged",
		GitCreatedAt: &now,
	}
	testCore.Create(unmerged)

	noBranch := &bean.Bean{
		ID:     "beans-no-branch",
		Slug:   "no-branch",
		Title:  "No Branch",
		Status: "todo",
	}
	testCore.Create(noBranch)

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	qr := resolver.Query()

	t.Run("filter gitBranchMerged=true", func(t *testing.T) {
		gitBranchMerged := true
		filter := &model.BeanFilter{
			GitBranchMerged: &gitBranchMerged,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 merged bean, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-merged" {
			t.Errorf("expected beans-merged, got %s", beans[0].ID)
		}
	})

	t.Run("filter gitBranchMerged=false", func(t *testing.T) {
		gitBranchMerged := false
		filter := &model.BeanFilter{
			GitBranchMerged: &gitBranchMerged,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		// Should include beans without branches AND beans with unmerged branches
		if len(beans) != 2 {
			t.Errorf("expected 2 beans (unmerged + no branch), got %d", len(beans))
		}
	})
}

func TestQueryCommand_CombinedGitAndStatusFilters(t *testing.T) {
	testCore, cleanup := setupQueryGitTestCore(t)
	defer cleanup()

	// Create beans with various combinations
	now := time.Now()

	bean1 := &bean.Bean{
		ID:           "beans-1",
		Slug:         "in-progress-with-branch",
		Title:        "In Progress With Branch",
		Status:       "in-progress",
		GitBranch:    "beans-1/branch",
		GitCreatedAt: &now,
	}
	testCore.Create(bean1)

	bean2 := &bean.Bean{
		ID:     "beans-2",
		Slug:   "in-progress-no-branch",
		Title:  "In Progress No Branch",
		Status: "in-progress",
	}
	testCore.Create(bean2)

	bean3 := &bean.Bean{
		ID:             "beans-3",
		Slug:           "completed-merged",
		Title:          "Completed Merged",
		Status:         "completed",
		GitBranch:      "beans-3/branch",
		GitMergedAt:    &now,
		GitMergeCommit: "abc123",
	}
	testCore.Create(bean3)

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	qr := resolver.Query()

	t.Run("in-progress with git branch", func(t *testing.T) {
		hasGitBranch := true
		filter := &model.BeanFilter{
			Status:       []string{"in-progress"},
			HasGitBranch: &hasGitBranch,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 bean, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-1" {
			t.Errorf("expected beans-1, got %s", beans[0].ID)
		}
	})

	t.Run("completed with merged branch", func(t *testing.T) {
		gitBranchMerged := true
		filter := &model.BeanFilter{
			Status:          []string{"completed"},
			GitBranchMerged: &gitBranchMerged,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 bean, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-3" {
			t.Errorf("expected beans-3, got %s", beans[0].ID)
		}
	})
}

func TestQueryCommand_GitFieldsInResult(t *testing.T) {
	testCore, cleanup := setupQueryGitTestCore(t)
	defer cleanup()

	// Create bean with full git metadata
	createdAt := time.Now().Add(-48 * time.Hour)
	mergedAt := time.Now()

	b := &bean.Bean{
		ID:             "beans-full-git",
		Slug:           "full-git",
		Title:          "Full Git Metadata",
		Status:         "completed",
		GitBranch:      "beans-full-git/full-git",
		GitCreatedAt:   &createdAt,
		GitMergedAt:    &mergedAt,
		GitMergeCommit: "abc123def456",
	}
	testCore.Create(b)

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	qr := resolver.Query()

	// Query the bean
	bean, err := qr.Bean(ctx, "beans-full-git")
	if err != nil {
		t.Fatalf("Bean() error = %v", err)
	}

	// Verify all git fields are present
	if bean.GitBranch != "beans-full-git/full-git" {
		t.Errorf("GitBranch = %q, want %q", bean.GitBranch, "beans-full-git/full-git")
	}
	if bean.GitCreatedAt == nil {
		t.Error("GitCreatedAt should not be nil")
	}
	if bean.GitMergedAt == nil {
		t.Error("GitMergedAt should not be nil")
	}
	if bean.GitMergeCommit != "abc123def456" {
		t.Errorf("GitMergeCommit = %q, want %q", bean.GitMergeCommit, "abc123def456")
	}
}

func TestQueryCommand_DocumentationExampleQueries(t *testing.T) {
	testCore, cleanup := setupQueryGitTestCore(t)
	defer cleanup()

	// Set up test data matching documentation examples
	now := time.Now()

	// Active bean with branch
	active := &bean.Bean{
		ID:           "beans-active",
		Slug:         "active-feature",
		Title:        "Active Feature",
		Status:       "in-progress",
		GitBranch:    "beans-active/active-feature",
		GitCreatedAt: &now,
	}
	testCore.Create(active)

	// Completed bean (merged)
	completed := &bean.Bean{
		ID:             "beans-completed",
		Slug:           "completed-feature",
		Title:          "Completed Feature",
		Status:         "completed",
		GitBranch:      "beans-completed/completed-feature",
		GitMergedAt:    &now,
		GitMergeCommit: "abc123",
	}
	testCore.Create(completed)

	// Todo bean (no branch yet)
	todo := &bean.Bean{
		ID:     "beans-todo",
		Slug:   "todo-task",
		Title:  "Todo Task",
		Status: "todo",
	}
	testCore.Create(todo)

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	qr := resolver.Query()

	// Example query 1: Get all beans with git branches
	t.Run("example: all beans with git branches", func(t *testing.T) {
		hasGitBranch := true
		filter := &model.BeanFilter{
			HasGitBranch: &hasGitBranch,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 2 {
			t.Errorf("expected 2 beans with branches, got %d", len(beans))
		}
	})

	// Example query 2: Get unmerged branches (active work)
	t.Run("example: unmerged branches (active work)", func(t *testing.T) {
		hasGitBranch := true
		gitBranchMerged := false
		filter := &model.BeanFilter{
			HasGitBranch:    &hasGitBranch,
			GitBranchMerged: &gitBranchMerged,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 unmerged branch, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-active" {
			t.Errorf("expected beans-active, got %s", beans[0].ID)
		}
	})

	// Example query 3: Get merged branches
	t.Run("example: merged branches", func(t *testing.T) {
		gitBranchMerged := true
		filter := &model.BeanFilter{
			GitBranchMerged: &gitBranchMerged,
		}

		beans, err := qr.Beans(ctx, filter)
		if err != nil {
			t.Fatalf("Beans() error = %v", err)
		}

		if len(beans) != 1 {
			t.Errorf("expected 1 merged branch, got %d", len(beans))
		}

		if len(beans) > 0 && beans[0].ID != "beans-completed" {
			t.Errorf("expected beans-completed, got %s", beans[0].ID)
		}
	})
}
