package gitflow

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestGetBranchStatus_Active(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a feature branch with unmerged commits
	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	featureRef := plumbing.NewBranchReferenceName("feature")
	repo.Storer.SetReference(plumbing.NewHashReference(featureRef, mainRef.Hash()))

	w, _ := repo.Worktree()
	w.Checkout(&git.CheckoutOptions{Branch: featureRef})
	commitFile(t, repo, "feature.txt", "feature content", "Add feature")

	// Switch back to main
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})

	status, err := gf.GetBranchStatus("feature", "main")
	if err != nil {
		t.Fatalf("GetBranchStatus() error = %v", err)
	}
	if status != BranchStatusActive {
		t.Errorf("GetBranchStatus() = %v, want %v (BranchStatusActive)", status, BranchStatusActive)
	}
}

func TestGetBranchStatus_Merged(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a feature branch
	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	featureRef := plumbing.NewBranchReferenceName("feature")
	repo.Storer.SetReference(plumbing.NewHashReference(featureRef, mainRef.Hash()))

	w, _ := repo.Worktree()
	w.Checkout(&git.CheckoutOptions{Branch: featureRef})
	featureCommit := commitFile(t, repo, "feature.txt", "feature content", "Add feature")

	// Merge to main (fast-forward)
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})
	repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), featureCommit))

	status, err := gf.GetBranchStatus("feature", "main")
	if err != nil {
		t.Fatalf("GetBranchStatus() error = %v", err)
	}
	if status != BranchStatusMerged {
		t.Errorf("GetBranchStatus() = %v, want %v (BranchStatusMerged)", status, BranchStatusMerged)
	}
}

func TestGetBranchStatus_Deleted(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Create a feature branch
	mainRef, _ := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	featureRef := plumbing.NewBranchReferenceName("feature")
	repo.Storer.SetReference(plumbing.NewHashReference(featureRef, mainRef.Hash()))

	w, _ := repo.Worktree()
	w.Checkout(&git.CheckoutOptions{Branch: featureRef})
	commitFile(t, repo, "feature.txt", "feature content", "Add feature")

	// Switch back to main
	w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})

	// Delete the branch without merging
	repo.Storer.RemoveReference(featureRef)

	status, err := gf.GetBranchStatus("feature", "main")
	if err != nil {
		t.Fatalf("GetBranchStatus() error = %v", err)
	}
	if status != BranchStatusDeleted {
		t.Errorf("GetBranchStatus() = %v, want %v (BranchStatusDeleted)", status, BranchStatusDeleted)
	}
}

func TestGetBranchStatus_NonexistentBranch(t *testing.T) {
	tmpDir, _ := setupTestRepo(t)
	gf, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Check status of a branch that never existed
	status, err := gf.GetBranchStatus("never-existed", "main")
	if err != nil {
		t.Fatalf("GetBranchStatus() error = %v", err)
	}
	// Should return Deleted since it doesn't exist and wasn't merged
	if status != BranchStatusDeleted {
		t.Errorf("GetBranchStatus() = %v, want %v (BranchStatusDeleted)", status, BranchStatusDeleted)
	}
}
