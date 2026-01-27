package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// setupSyncTestEnv creates a test environment with both a git repo and beans core
func setupSyncTestEnv(t *testing.T) (*beancore.Core, *git.Repository, string, func()) {
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

// createTestBranch creates a git branch and commits a file
func createTestBranch(t *testing.T, repo *git.Repository, branchName, fileName string) plumbing.Hash {
	t.Helper()

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create branch
	headRef, err := repo.Head()
	if err != nil {
		t.Fatalf("failed to get HEAD: %v", err)
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRef, headRef.Hash()))
	if err != nil {
		t.Fatalf("failed to create branch ref: %v", err)
	}

	// Check out the branch
	if err := w.Checkout(&git.CheckoutOptions{Branch: branchRef}); err != nil {
		t.Fatalf("failed to checkout branch: %v", err)
	}

	// Commit a file
	filePath := filepath.Join(w.Filesystem.Root(), fileName)
	if err := os.WriteFile(filePath, []byte("test content\n"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	w.Add(fileName)
	hash, err := w.Commit("Add "+fileName, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	return hash
}

// mergeToMain merges a branch to main (fast-forward)
func mergeToMain(t *testing.T, repo *git.Repository, commitHash plumbing.Hash) {
	t.Helper()

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Check out main
	mainRef := plumbing.NewBranchReferenceName("main")
	if err := w.Checkout(&git.CheckoutOptions{Branch: mainRef}); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Update main to point to the commit (fast-forward merge)
	if err := repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commitHash)); err != nil {
		t.Fatalf("failed to merge to main: %v", err)
	}
}

// deleteBranch deletes a git branch
func deleteBranch(t *testing.T, repo *git.Repository, branchName string) {
	t.Helper()

	branchRef := plumbing.NewBranchReferenceName(branchName)
	if err := repo.Storer.RemoveReference(branchRef); err != nil {
		t.Fatalf("failed to delete branch: %v", err)
	}
}

// commitAll commits all changes in the working tree
func commitAll(t *testing.T, repo *git.Repository, message string) {
	t.Helper()

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Add all changes
	if err := w.AddGlob("."); err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	// Commit
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestSyncCommand_NoBeansWithGitBranches(t *testing.T) {
	testCore, _, _, cleanup := setupSyncTestEnv(t)
	defer cleanup()

	// Create a bean without git branch
	b := &bean.Bean{
		ID:     "test-1",
		Slug:   "test-bean",
		Title:  "Test Bean",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("failed to create bean: %v", err)
	}

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}

	// Call sync mutation
	updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
	if err != nil {
		t.Fatalf("SyncGitBranches() error = %v", err)
	}

	// Should return empty list since no beans have git branches
	if len(updatedBeans) != 0 {
		t.Errorf("expected 0 beans to sync, got %d", len(updatedBeans))
	}
}

func TestSyncCommand_MergedBranch(t *testing.T) {
	testCore, repo, tmpDir, cleanup := setupSyncTestEnv(t)
	defer cleanup()

	// Create parent bean
	parent := &bean.Bean{
		ID:     "parent-1",
		Slug:   "parent-bean",
		Title:  "Parent Bean",
		Status: "todo",
		Type:   "epic",
	}
	if err := testCore.Create(parent); err != nil {
		t.Fatalf("failed to create parent bean: %v", err)
	}

	// Create child to make it a parent
	child := &bean.Bean{
		ID:     "child-1",
		Slug:   "child-bean",
		Title:  "Child Bean",
		Status: "todo",
		Type:   "task",
		Parent: "parent-1",
	}
	if err := testCore.Create(child); err != nil {
		t.Fatalf("failed to create child bean: %v", err)
	}

	// Commit beans to git (required for clean working tree)
	commitAll(t, repo, "Add beans")

	// Transition to in-progress should create git branch
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	_, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Reload parent to get git branch info
	updatedParent, err := testCore.Get(parent.ID)
	if err != nil {
		t.Fatalf("failed to reload parent: %v", err)
	}
	if updatedParent.GitBranch == "" {
		t.Fatal("expected git branch to be created")
	}

	// The gitflow already created and checked out the branch
	// Commit the bean updates first
	commitAll(t, repo, "Update bean with git branch")

	// Add a commit on the feature branch
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	featureFile := filepath.Join(tmpDir, "feature.txt")
	if err := os.WriteFile(featureFile, []byte("feature content\n"), 0644); err != nil {
		t.Fatalf("failed to write feature file: %v", err)
	}
	w.Add("feature.txt")
	commitHash, err := w.Commit("Add feature", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit feature: %v", err)
	}

	// Merge to main (switch to main and fast-forward)
	mergeToMain(t, repo, commitHash)

	// Now sync should mark the bean as completed
	updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
	if err != nil {
		t.Fatalf("SyncGitBranches() error = %v", err)
	}

	if len(updatedBeans) != 1 {
		t.Fatalf("expected 1 bean to sync, got %d", len(updatedBeans))
	}

	if updatedBeans[0].Status != "completed" {
		t.Errorf("expected status 'completed', got %q", updatedBeans[0].Status)
	}
}

func TestSyncCommand_DeletedBranch(t *testing.T) {
	testCore, repo, tmpDir, cleanup := setupSyncTestEnv(t)
	defer cleanup()

	// Create parent bean with child
	parent := &bean.Bean{
		ID:     "parent-1",
		Slug:   "parent-bean",
		Title:  "Parent Bean",
		Status: "todo",
		Type:   "epic",
	}
	if err := testCore.Create(parent); err != nil {
		t.Fatalf("failed to create parent bean: %v", err)
	}

	child := &bean.Bean{
		ID:     "child-1",
		Slug:   "child-bean",
		Title:  "Child Bean",
		Status: "todo",
		Type:   "task",
		Parent: "parent-1",
	}
	if err := testCore.Create(child); err != nil {
		t.Fatalf("failed to create child bean: %v", err)
	}

	// Commit beans to git (required for clean working tree)
	commitAll(t, repo, "Add beans")

	// Transition to in-progress to create git branch
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}
	_, err := resolver.Mutation().UpdateBean(ctx, parent.ID, model.UpdateBeanInput{
		Status: stringPtr("in-progress"),
	})
	if err != nil {
		t.Fatalf("UpdateBean() error = %v", err)
	}

	// Reload parent to get git branch info
	updatedParent, err := testCore.Get(parent.ID)
	if err != nil {
		t.Fatalf("failed to reload parent: %v", err)
	}
	if updatedParent.GitBranch == "" {
		t.Fatal("expected git branch to be created")
	}

	// The gitflow already created and checked out the branch
	// Commit the bean updates first
	commitAll(t, repo, "Update bean with git branch")

	// Add a commit on the feature branch
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	featureFile := filepath.Join(tmpDir, "feature.txt")
	if err := os.WriteFile(featureFile, []byte("feature content\n"), 0644); err != nil {
		t.Fatalf("failed to write feature file: %v", err)
	}
	w.Add("feature.txt")
	_, err = w.Commit("Add feature", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit feature: %v", err)
	}

	// Switch back to main before deleting the branch
	if err := w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")}); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Delete the branch (without merging)
	deleteBranch(t, repo, updatedParent.GitBranch)

	// Now sync should mark the bean as scrapped
	updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
	if err != nil {
		t.Fatalf("SyncGitBranches() error = %v", err)
	}

	if len(updatedBeans) != 1 {
		t.Fatalf("expected 1 bean to sync, got %d", len(updatedBeans))
	}

	if updatedBeans[0].Status != "scrapped" {
		t.Errorf("expected status 'scrapped', got %q", updatedBeans[0].Status)
	}
}

func TestSyncCommand_JSONOutput(t *testing.T) {
	testCore, _, _, cleanup := setupSyncTestEnv(t)
	defer cleanup()

	// Create a bean without git branch
	b := &bean.Bean{
		ID:     "test-1",
		Slug:   "test-bean",
		Title:  "Test Bean",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("failed to create bean: %v", err)
	}

	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}

	// Call sync mutation
	updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
	if err != nil {
		t.Fatalf("SyncGitBranches() error = %v", err)
	}

	// Test JSON output formatting
	response := output.Response{
		Success: true,
		Beans:   updatedBeans,
		Count:   len(updatedBeans),
		Message: "Git branches synced",
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	// Verify it's valid JSON
	var parsed output.Response
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if !parsed.Success {
		t.Error("expected success to be true")
	}
	if parsed.Message != "Git branches synced" {
		t.Errorf("expected message 'Git branches synced', got %q", parsed.Message)
	}
	if parsed.Count != 0 {
		t.Errorf("expected count 0, got %d", parsed.Count)
	}
}

func TestSyncCommand_MixedStates(t *testing.T) {
	testCore, repo, tmpDir, cleanup := setupSyncTestEnv(t)
	defer cleanup()

	// Create multiple parent beans with different branch states
	ctx := context.Background()
	resolver := &graph.Resolver{Core: testCore}

	// Bean 1: Will be merged
	parent1 := &bean.Bean{
		ID:     "parent-1",
		Slug:   "merged-bean",
		Title:  "Merged Bean",
		Status: "todo",
		Type:   "epic",
	}
	child1 := &bean.Bean{
		ID:     "child-1",
		Slug:   "child-1",
		Title:  "Child 1",
		Status: "todo",
		Type:   "task",
		Parent: "parent-1",
	}
	testCore.Create(parent1)
	testCore.Create(child1)

	// Bean 2: Will be deleted
	parent2 := &bean.Bean{
		ID:     "parent-2",
		Slug:   "deleted-bean",
		Title:  "Deleted Bean",
		Status: "todo",
		Type:   "epic",
	}
	child2 := &bean.Bean{
		ID:     "child-2",
		Slug:   "child-2",
		Title:  "Child 2",
		Status: "todo",
		Type:   "task",
		Parent: "parent-2",
	}
	testCore.Create(parent2)
	testCore.Create(child2)

	// Bean 3: Will remain active (no branch operations)
	parent3 := &bean.Bean{
		ID:     "parent-3",
		Slug:   "active-bean",
		Title:  "Active Bean",
		Status: "todo",
		Type:   "epic",
	}
	child3 := &bean.Bean{
		ID:     "child-3",
		Slug:   "child-3",
		Title:  "Child 3",
		Status: "todo",
		Type:   "task",
		Parent: "parent-3",
	}
	testCore.Create(parent3)
	testCore.Create(child3)

	// Commit beans to git (required for clean working tree)
	commitAll(t, repo, "Add beans")

	// Transition all to in-progress to create branches
	for _, id := range []string{"parent-1", "parent-2", "parent-3"} {
		_, err := resolver.Mutation().UpdateBean(ctx, id, model.UpdateBeanInput{
			Status: stringPtr("in-progress"),
		})
		if err != nil {
			t.Fatalf("UpdateBean(%s) error = %v", id, err)
		}
		// Commit bean updates after each transition
		commitAll(t, repo, fmt.Sprintf("Update %s to in-progress", id))
	}

	// Reload to get branch names
	p1, _ := testCore.Get("parent-1")
	p2, _ := testCore.Get("parent-2")
	p3, _ := testCore.Get("parent-3")

	// At this point we're on parent-3's branch (the last one created)
	// We need to work with each branch individually

	// Process parent-1: checkout, commit, merge
	w, _ := repo.Worktree()
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(p1.GitBranch)})
	f1 := filepath.Join(tmpDir, "feature1.txt")
	os.WriteFile(f1, []byte("feature 1\n"), 0644)
	w.Add("feature1.txt")
	hash1, _ := w.Commit("Add feature 1", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	mergeToMain(t, repo, hash1)

	// Process parent-2: checkout, commit, then delete without merging
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(p2.GitBranch)})
	f2 := filepath.Join(tmpDir, "feature2.txt")
	os.WriteFile(f2, []byte("feature 2\n"), 0644)
	w.Add("feature2.txt")
	w.Commit("Add feature 2", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})
	deleteBranch(t, repo, p2.GitBranch)

	// Parent-3's branch stays active (has commits but not merged or deleted)
	// We can add a commit to it to make it realistic
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(p3.GitBranch)})
	f3 := filepath.Join(tmpDir, "feature3.txt")
	os.WriteFile(f3, []byte("feature 3\n"), 0644)
	w.Add("feature3.txt")
	w.Commit("Add feature 3", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	// Stay on this branch (it's still active)

	// Sync should update parent-1 and parent-2, but not parent-3
	updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
	if err != nil {
		t.Fatalf("SyncGitBranches() error = %v", err)
	}

	// Should have synced 2 beans (merged and deleted)
	if len(updatedBeans) != 2 {
		t.Fatalf("expected 2 beans to sync, got %d", len(updatedBeans))
	}

	// Check statuses
	statusMap := make(map[string]string)
	for _, b := range updatedBeans {
		statusMap[b.ID] = b.Status
	}

	if statusMap["parent-1"] != "completed" {
		t.Errorf("expected parent-1 to be completed, got %q", statusMap["parent-1"])
	}
	if statusMap["parent-2"] != "scrapped" {
		t.Errorf("expected parent-2 to be scrapped, got %q", statusMap["parent-2"])
	}

	// parent-3 should still be in-progress
	p3Updated, _ := testCore.Get("parent-3")
	if p3Updated.Status != "in-progress" {
		t.Errorf("expected parent-3 to remain in-progress, got %q", p3Updated.Status)
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
