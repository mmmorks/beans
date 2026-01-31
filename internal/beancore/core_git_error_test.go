package beancore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/config"
)

// setupGitTestCore creates a test core with git integration
func setupGitTestCore(t *testing.T, cfg *config.Config) (*Core, *git.Repository, string) {
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

	// Create core with provided config
	if cfg == nil {
		cfg = config.Default()
		cfg.Beans.Git.Enabled = true
		cfg.Beans.Git.AutoCreateBranch = true
		cfg.Beans.Git.BaseBranch = "main"
	}

	testCore := New(beansDir, cfg)
	testCore.SetWarnWriter(nil)
	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// Enable git integration
	if cfg.Beans.Git.Enabled {
		if err := testCore.EnableGitFlow(tmpDir); err != nil {
			t.Fatalf("failed to enable gitflow: %v", err)
		}
	}

	return testCore, repo, tmpDir
}

func TestCore_GitBranchCreation_RequireMerge(t *testing.T) {
	t.Skip("RequireMerge feature not yet implemented")

	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	cfg.Beans.Git.AutoCreateBranch = true
	cfg.Beans.Git.BaseBranch = "main"
	cfg.Beans.Git.RequireMerge = true

	testCore, repo, _ := setupGitTestCore(t, cfg)

	// Create parent bean with child
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

	// Commit beans to git
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Transition to in-progress (creates branch)
	parent.Status = "in-progress"
	if err := testCore.Update(parent, nil); err != nil {
		t.Fatalf("Update to in-progress error = %v", err)
	}

	// Commit the bean update
	w.Add(".beans")
	w.Commit("Update to in-progress", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Attempt to manually complete without merging branch (should fail)
	parent.Status = "completed"
	err := testCore.Update(parent, nil)
	if err == nil {
		t.Error("Update to completed should fail when require_merge is true and branch not merged")
	}
	if err != nil && !contains(err.Error(), "merge") && !contains(err.Error(), "branch") {
		t.Errorf("error should mention merge/branch requirement, got: %v", err)
	}
}

func TestCore_GitBranchCreation_AutoCommitBeans(t *testing.T) {
	t.Skip("Auto-commit during branch creation has different behavior - tested separately in core_test.go")

	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	cfg.Beans.Git.AutoCreateBranch = true
	cfg.Beans.Git.BaseBranch = "main"
	cfg.Beans.Git.AutoCommitBeans = true

	testCore, repo, _ := setupGitTestCore(t, cfg)

	// Create parent bean with child
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

	// Commit beans first (auto-commit happens during branch creation if working tree is clean)
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Transition to in-progress - should create branch and auto-commit bean updates
	parent.Status = "in-progress"
	if err := testCore.Update(parent, nil); err != nil {
		t.Fatalf("Update to in-progress error = %v", err)
	}

	// Verify working tree is clean (bean update was auto-committed)
	status, _ := w.Status()
	if !status.IsClean() {
		t.Error("working tree should be clean after auto-commit")
	}
}

func TestCore_GitBranchCreation_NoAutoCommit(t *testing.T) {
	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	cfg.Beans.Git.AutoCreateBranch = true
	cfg.Beans.Git.BaseBranch = "main"
	cfg.Beans.Git.AutoCommitBeans = false

	testCore, repo, _ := setupGitTestCore(t, cfg)

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

	// Commit beans manually
	w, _ := repo.Worktree()
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Transition to in-progress (creates branch but does NOT auto-commit)
	parent.Status = "in-progress"
	if err := testCore.Update(parent, nil); err != nil {
		t.Fatalf("Update to in-progress error = %v", err)
	}

	// Working tree should have uncommitted changes to bean files
	status, _ := w.Status()
	if status.IsClean() {
		t.Error("working tree should have uncommitted bean changes when auto-commit is disabled")
	}
}

func TestCore_GitBranchCreation_InvalidBaseBranch(t *testing.T) {
	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	cfg.Beans.Git.AutoCreateBranch = true
	cfg.Beans.Git.BaseBranch = "nonexistent-branch"

	testCore, repo, _ := setupGitTestCore(t, cfg)

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

	// Attempt to transition to in-progress (should fail due to invalid base branch)
	parent.Status = "in-progress"
	err := testCore.Update(parent, nil)
	if err == nil {
		t.Error("Update should fail with invalid base branch configured")
	}
	if !contains(err.Error(), "not found") && !contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention base branch not found, got: %v", err)
	}
}

func TestCore_GitIntegration_AutoEnable(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create initial commit
	w, _ := repo.Worktree()
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test\n"), 0644)
	w.Add("README.md")
	commit, _ := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	mainRef := plumbing.NewBranchReferenceName("main")
	repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commit))
	w.Checkout(&git.CheckoutOptions{Branch: mainRef})

	// Create beans directory
	beansDir := filepath.Join(tmpDir, ".beans")
	os.MkdirAll(beansDir, 0755)

	// Create core with git enabled but not explicitly enabled yet
	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	testCore := New(beansDir, cfg)
	testCore.SetWarnWriter(nil)

	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// Git flow should not be enabled yet
	if testCore.GitFlow() != nil {
		t.Error("gitflow should not be enabled before EnableGitFlow is called")
	}

	// Auto-enable git integration
	if err := testCore.EnableGitFlow(tmpDir); err != nil {
		t.Fatalf("EnableGitFlow() error = %v", err)
	}

	// Now gitflow should be enabled
	if testCore.GitFlow() == nil {
		t.Error("gitflow should be enabled after EnableGitFlow")
	}
}

func TestCore_GitIntegration_DisabledInConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create initial commit
	w, _ := repo.Worktree()
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Test\n"), 0644)
	w.Add("README.md")
	commit, _ := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	mainRef := plumbing.NewBranchReferenceName("main")
	repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commit))
	w.Checkout(&git.CheckoutOptions{Branch: mainRef})

	// Create beans directory
	beansDir := filepath.Join(tmpDir, ".beans")
	os.MkdirAll(beansDir, 0755)

	// Create core with git explicitly disabled
	cfg := config.Default()
	cfg.Beans.Git.Enabled = false
	cfg.Beans.Git.AutoCreateBranch = false
	testCore := New(beansDir, cfg)
	testCore.SetWarnWriter(nil)

	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// DON'T call EnableGitFlow when git is disabled in config
	// The core should respect the config setting

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
	w.Add(".beans")
	w.Commit("Add beans", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	// Transition to in-progress (should NOT create git branch)
	parent.Status = "in-progress"
	if err := testCore.Update(parent, nil); err != nil {
		t.Fatalf("Update should succeed even with git disabled: %v", err)
	}

	// Reload and verify no git branch was created
	updated, _ := testCore.Get(parent.ID)
	if updated.GitBranch != "" {
		t.Errorf("GitBranch should be empty when git is disabled, got %q", updated.GitBranch)
	}
}

func TestCore_GitIntegration_NoGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create beans directory WITHOUT initializing git
	beansDir := filepath.Join(tmpDir, ".beans")
	os.MkdirAll(beansDir, 0755)

	// Create core with git enabled
	cfg := config.Default()
	cfg.Beans.Git.Enabled = true
	testCore := New(beansDir, cfg)
	testCore.SetWarnWriter(nil)

	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// Attempt to enable git (should fail since no .git directory)
	err := testCore.EnableGitFlow(tmpDir)
	if err == nil {
		t.Error("EnableGitFlow() should fail when no git repository exists")
	}

	// Gitflow should remain nil
	if testCore.GitFlow() != nil {
		t.Error("gitflow should be nil when EnableGitFlow fails")
	}

	// Operations should still work without git
	bean := &bean.Bean{
		ID:     "test-1",
		Slug:   "test-bean",
		Title:  "Test Bean",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(bean); err != nil {
		t.Fatalf("Create should work without git: %v", err)
	}
}

// Note: contains() helper function is defined in core_test.go
