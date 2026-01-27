package gitflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestCreateBranch_DetachedHEAD(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a commit and detach HEAD
	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	commit, _ := repo.CommitObject(mainRef.Hash())

	w, _ := repo.Worktree()
	err = w.Checkout(&git.CheckoutOptions{
		Hash: commit.Hash,
	})
	if err != nil {
		t.Fatalf("failed to checkout detached HEAD: %v", err)
	}

	// Attempt to create a branch from detached HEAD state
	_, err = gf.CreateBranch("beans-test", "feature", "main")
	// Should succeed because we specify the base branch
	if err != nil {
		t.Errorf("CreateBranch() in detached HEAD should succeed: %v", err)
	}
}

func TestCreateBranch_InvalidBaseBranch(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = gf.CreateBranch("beans-test", "feature", "nonexistent-base")
	if err == nil {
		t.Error("CreateBranch() with invalid base branch should return error")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestGetMainBranch_NoRemoteHead(t *testing.T) {
	// Create a repo without any remote
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create initial commit on main
	w, _ := repo.Worktree()
	testFile := filepath.Join(tmpDir, "README.md")
	os.WriteFile(testFile, []byte("# Test\n"), 0644)
	w.Add("README.md")
	commit, _ := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	mainRef := plumbing.NewBranchReferenceName("main")
	repo.Storer.SetReference(plumbing.NewHashReference(mainRef, commit))
	w.Checkout(&git.CheckoutOptions{Branch: mainRef})

	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should fall back to detecting "main" locally
	branch, err := gf.GetMainBranch()
	if err != nil {
		t.Fatalf("GetMainBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("GetMainBranch() = %q, want %q", branch, "main")
	}
}

func TestGetMainBranch_OnlyMasterExists(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create initial commit on master (not main)
	w, _ := repo.Worktree()
	testFile := filepath.Join(tmpDir, "README.md")
	os.WriteFile(testFile, []byte("# Test\n"), 0644)
	w.Add("README.md")
	commit, _ := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})

	masterRef := plumbing.NewBranchReferenceName("master")
	repo.Storer.SetReference(plumbing.NewHashReference(masterRef, commit))
	w.Checkout(&git.CheckoutOptions{Branch: masterRef})

	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should detect "master" as the main branch
	branch, err := gf.GetMainBranch()
	if err != nil {
		t.Fatalf("GetMainBranch() error = %v", err)
	}
	if branch != "master" {
		t.Errorf("GetMainBranch() = %q, want %q", branch, "master")
	}
}

func TestGetMainBranch_NoBranches(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create gitflow without any commits or branches
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should return an error when no branches exist
	_, err = gf.GetMainBranch()
	if err == nil {
		t.Error("GetMainBranch() should fail when no branches exist")
	}
}

func TestIsWorkingTreeClean_UnstagedChanges(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Modify an existing file without staging
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Modified\n"), 0644)

	clean, err := gf.IsWorkingTreeClean()
	if err != nil {
		t.Fatalf("IsWorkingTreeClean() error = %v", err)
	}
	if clean {
		t.Error("IsWorkingTreeClean() = true, want false (unstaged changes)")
	}
}

func TestIsWorkingTreeClean_StagedChanges(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add a new file and stage it
	newFile := filepath.Join(tmpDir, "new-file.txt")
	os.WriteFile(newFile, []byte("new content"), 0644)

	w, _ := repo.Worktree()
	w.Add("new-file.txt")

	clean, err := gf.IsWorkingTreeClean()
	if err != nil {
		t.Fatalf("IsWorkingTreeClean() error = %v", err)
	}
	if clean {
		t.Error("IsWorkingTreeClean() = true, want false (staged changes)")
	}
}

func TestCommitBeans_NoChanges(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create .beans directory but don't add any files
	beansDir := filepath.Join(tmpDir, ".beans")
	os.MkdirAll(beansDir, 0755)

	// Commit with no .beans files should fail
	err = gf.CommitBeans("test: no changes")
	if err == nil {
		t.Error("CommitBeans() with no .beans files should return error")
	}
	if !contains(err.Error(), "glob") && !contains(err.Error(), "match") {
		t.Errorf("error should mention glob/match issue, got: %v", err)
	}
}

func TestCommitBeans_OnlyBeansChanges(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create .beans directory and add files
	beansDir := filepath.Join(tmpDir, ".beans")
	os.MkdirAll(beansDir, 0755)
	beanFile := filepath.Join(beansDir, "bean-1.md")
	os.WriteFile(beanFile, []byte("# Bean 1"), 0644)

	// Also add a non-beans file
	otherFile := filepath.Join(tmpDir, "other.txt")
	os.WriteFile(otherFile, []byte("other content"), 0644)

	// CommitBeans should only commit .beans files
	err = gf.CommitBeans("test: add bean")
	if err != nil {
		t.Fatalf("CommitBeans() error = %v", err)
	}

	// Verify other.txt is still uncommitted
	w, _ := repo.Worktree()
	status, _ := w.Status()
	if !status.IsUntracked("other.txt") {
		t.Error("other.txt should still be untracked after CommitBeans()")
	}

	// Verify bean file was committed
	if status.IsUntracked(".beans/bean-1.md") {
		t.Error(".beans/bean-1.md should be committed")
	}
}

func TestBranchExists_EmptyBranchName(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// BranchExists with empty name should return false (doesn't exist)
	exists, err := gf.BranchExists("")
	// The implementation may or may not return an error, but it should return false
	if exists {
		t.Error("BranchExists(\"\") should return false")
	}
}

func TestSwitchBranch_NonexistentBranch(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = gf.SwitchBranch("nonexistent")
	if err == nil {
		t.Error("SwitchBranch() to nonexistent branch should return error")
	}
}

func TestSwitchBranch_DirtyWorkingTree(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create another branch
	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	otherRef := plumbing.NewBranchReferenceName("other")
	repo.Storer.SetReference(plumbing.NewHashReference(otherRef, mainRef.Hash()))

	// Modify an existing tracked file (not just add untracked file)
	readmePath := filepath.Join(tmpDir, "README.md")
	os.WriteFile(readmePath, []byte("# Modified\n"), 0644)

	// Switch should fail with modified tracked file
	err = gf.SwitchBranch("other")
	if err == nil {
		t.Error("SwitchBranch() with dirty tree (modified tracked file) should return error")
	}
}

func TestIsBranchMerged_InvalidBranchNames(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Test with nonexistent branch name
	merged, _, err := gf.IsBranchMerged("nonexistent-branch", "main")
	// The implementation may not error, but it should return not merged
	if err != nil {
		t.Logf("IsBranchMerged with nonexistent branch returned error (acceptable): %v", err)
	}
	if merged {
		t.Error("IsBranchMerged for nonexistent branch should return false")
	}

	// Test with nonexistent base branch
	_, _, err = gf.IsBranchMerged("main", "nonexistent-base")
	// Should return an error since base branch doesn't exist
	if err == nil {
		t.Error("IsBranchMerged with nonexistent base branch should return error")
	}
}
