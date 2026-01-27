package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/graph/model"
	"github.com/hmans/beans/internal/output"
)

// setupUpdateTestEnv creates a test environment similar to sync tests
func setupUpdateTestEnv(t *testing.T) (*beancore.Core, *git.Repository, string, func()) {
	t.Helper()
	tmpDir := t.TempDir()

	// Initialize git repo
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create initial commit on main branch
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create a README file for initial commit
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}

	if _, err := w.Add("README.md"); err != nil {
		t.Fatalf("failed to add README: %v", err)
	}

	commit, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	// Create main branch reference
	mainRef := plumbing.NewBranchReferenceName("main")
	if err := repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commit)); err != nil {
		t.Fatalf("failed to create main branch: %v", err)
	}

	// Checkout main
	if err := w.Checkout(&git.CheckoutOptions{Branch: mainRef}); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Set up beans directory
	beansDir := filepath.Join(tmpDir, ".beans")
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatalf("failed to create .beans dir: %v", err)
	}

	// Create core with git integration enabled
	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	cfg.Beans.Git.AutoCreateBranch = true
	cfg.Beans.Git.BaseBranch = "main"
	testCore := beancore.New(beansDir, cfg)
	testCore.SetWarnWriter(nil) // suppress warnings in tests
	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// Enable git integration
	if err := testCore.EnableGitFlow(tmpDir); err != nil {
		t.Fatalf("failed to enable gitflow: %v", err)
	}

	// Save and restore the global core
	oldCore := core
	core = testCore

	cleanup := func() {
		core = oldCore
	}

	return testCore, repo, tmpDir, cleanup
}

// setupUpdateTestEnvNoGit creates a test environment without git integration
func setupUpdateTestEnvNoGit(t *testing.T) (*beancore.Core, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	beansDir := filepath.Join(tmpDir, ".beans")
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

func TestUpdateCommand_GitBranchAutoCreate_ParentBean(t *testing.T) {
	testCore, repo, _, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create parent bean with a child
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	if err := testCore.Create(parent); err != nil {
		t.Fatalf("Create parent error = %v", err)
	}

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	if err := testCore.Create(child); err != nil {
		t.Fatalf("Create child error = %v", err)
	}

	// Commit beans to git (working tree must be clean)
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Update parent to in-progress via GraphQL mutation (simulating CLI)
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	updated, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Verify git branch was created
	expectedBranch := "beans-parent1/parent-feature"
	if updated.GitBranch != expectedBranch {
		t.Errorf("GitBranch = %q, want %q", updated.GitBranch, expectedBranch)
	}

	if updated.GitCreatedAt == nil {
		t.Error("GitCreatedAt should be set")
	}

	// Verify branch exists in git
	branchRef := plumbing.NewBranchReferenceName(expectedBranch)
	_, err = repo.Reference(branchRef, true)
	if err != nil {
		t.Errorf("git branch %q should exist: %v", expectedBranch, err)
	}

	// Verify we're on the new branch
	head, _ := repo.Head()
	if head.Name().Short() != expectedBranch {
		t.Errorf("current branch = %q, want %q", head.Name().Short(), expectedBranch)
	}
}

func TestUpdateCommand_GitBranchNoAutoCreate_NonParentBean(t *testing.T) {
	testCore, repo, _, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create a non-parent bean
	nonParent := &bean.Bean{
		ID:     "beans-solo1",
		Slug:   "solo-task",
		Title:  "Solo Task",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(nonParent); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Commit bean to git
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add bean", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Update to in-progress
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	updated, err := resolver.Mutation().UpdateBean(ctx, nonParent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Verify git branch was NOT created
	if updated.GitBranch != "" {
		t.Errorf("GitBranch = %q, want empty (non-parent should not get branch)", updated.GitBranch)
	}
}

func TestUpdateCommand_GitDisabled(t *testing.T) {
	testCore, cleanup := setupUpdateTestEnvNoGit(t)
	defer cleanup()

	// Create parent bean with child
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	testCore.Create(parent)

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	testCore.Create(child)

	// Update to in-progress
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	updated, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Verify git branch was NOT created (git disabled)
	if updated.GitBranch != "" {
		t.Errorf("GitBranch = %q, want empty (git disabled)", updated.GitBranch)
	}
}

func TestUpdateCommand_DirtyWorkingTree(t *testing.T) {
	testCore, repo, tmpDir, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create parent bean with child
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	testCore.Create(parent)

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	testCore.Create(child)

	// Commit beans
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Create uncommitted change
	dirtyFile := filepath.Join(tmpDir, "dirty.txt")
	os.WriteFile(dirtyFile, []byte("uncommitted\n"), 0644)

	// Attempt to update to in-progress (should fail with dirty tree)
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	_, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})

	// Should error due to dirty working tree
	if err == nil {
		t.Error("UpdateBean() should fail with dirty working tree")
	}
	if !strings.Contains(err.Error(), "dirty") && !strings.Contains(err.Error(), "uncommitted") {
		t.Errorf("Error should mention dirty/uncommitted changes, got: %v", err)
	}
}

func TestUpdateCommand_MultipleStatusTransitions(t *testing.T) {
	testCore, repo, _, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create parent bean with child
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	testCore.Create(parent)

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	testCore.Create(child)

	// Commit beans
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}

	// First transition: todo -> in-progress (creates branch)
	updated, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("First UpdateBean() error = %v", err)
	}
	if updated.GitBranch == "" {
		t.Error("GitBranch should be set after first transition")
	}

	// Commit the bean update
	w.Add(".beans")
	w.Commit("Update to in-progress", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Second transition: in-progress -> todo (branch should remain)
	updated, err = resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("todo"),
	})
	if err != nil {
		t.Fatalf("Second UpdateBean() error = %v", err)
	}
	if updated.GitBranch == "" {
		t.Error("GitBranch should still be set after second transition")
	}

	// Third transition: todo -> in-progress again (should not create new branch)
	w.Add(".beans")
	w.Commit("Update to todo", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	originalBranch := updated.GitBranch
	updated, err = resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("Third UpdateBean() error = %v", err)
	}
	if updated.GitBranch != originalBranch {
		t.Errorf("GitBranch changed after re-transitioning to in-progress: was %q, now %q", originalBranch, updated.GitBranch)
	}
}

func TestUpdateCommand_JSONOutput(t *testing.T) {
	testCore, repo, _, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create parent bean with child
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	testCore.Create(parent)

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	testCore.Create(child)

	// Commit beans
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Update to in-progress
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	updated, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Test JSON output includes git fields
	jsonBytes, err := json.Marshal(updated)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Parse back to verify fields
	var parsed bean.Bean
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if parsed.GitBranch == "" {
		t.Error("JSON output should include gitBranch field")
	}
	if parsed.GitBranch != updated.GitBranch {
		t.Errorf("JSON gitBranch = %q, want %q", parsed.GitBranch, updated.GitBranch)
	}
	if parsed.GitCreatedAt == nil {
		t.Error("JSON output should include gitCreatedAt field")
	}
}

func TestUpdateCommand_ResponseOutput(t *testing.T) {
	testCore, repo, _, cleanup := setupUpdateTestEnv(t)
	defer cleanup()

	// Create parent bean
	parent := &bean.Bean{
		ID:     "beans-parent1",
		Slug:   "parent-feature",
		Title:  "Parent Feature",
		Status: "todo",
		Type:   "epic",
	}
	testCore.Create(parent)

	child := &bean.Bean{
		ID:     "beans-child1",
		Slug:   "child-task",
		Title:  "Child Task",
		Status: "todo",
		Type:   "task",
		Parent: "beans-parent1",
	}
	testCore.Create(child)

	// Commit beans
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Update bean
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	updated, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Create output response (simulating CLI output)
	response := output.Response{
		Success: true,
		Bean:    updated,
		Message: "Updated " + updated.ID,
	}

	// Verify response can be marshaled
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Verify git fields are included in the bean object
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "git_branch") && !strings.Contains(jsonStr, "gitBranch") {
		t.Logf("JSON output: %s", jsonStr)
		t.Error("Response should include git branch field in output")
	}

	// Also verify the bean has the expected git branch value
	if updated.GitBranch == "" {
		t.Error("Updated bean should have GitBranch set")
	}
}
