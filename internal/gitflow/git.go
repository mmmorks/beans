package gitflow

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitFlow provides git operations for beans integration.
type GitFlow struct {
	repoPath string
	repo     *git.Repository
}

// New creates a new GitFlow instance for the given repository path.
// Returns an error if the path is not a git repository.
func New(repoPath string) (*GitFlow, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at %s: %w", repoPath, err)
	}

	return &GitFlow{
		repoPath: repoPath,
		repo:     repo,
	}, nil
}

// CreateBranch creates a new git branch with the given name from the current HEAD.
// Returns the full branch name and switches to it.
func (g *GitFlow) CreateBranch(beanID, slug string) (string, error) {
	branchName := BuildBranchName(beanID, slug)

	// Check if branch already exists
	exists, err := g.BranchExists(branchName)
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		return "", fmt.Errorf("branch %q already exists", branchName)
	}

	// Get HEAD reference
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch reference
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRefName, head.Hash())

	err = g.repo.Storer.SetReference(ref)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	// Switch to the new branch
	if err := g.SwitchBranch(branchName); err != nil {
		return "", fmt.Errorf("failed to switch to new branch: %w", err)
	}

	return branchName, nil
}

// GetCurrentBranch returns the name of the currently checked out branch.
// Returns an error if in detached HEAD state.
func (g *GitFlow) GetCurrentBranch() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("not on a branch (detached HEAD)")
	}

	return head.Name().Short(), nil
}

// BranchExists checks if a branch with the given name exists locally.
func (g *GitFlow) BranchExists(branchName string) (bool, error) {
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	_, err := g.repo.Reference(branchRefName, true)
	if err == plumbing.ErrReferenceNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check branch reference: %w", err)
	}
	return true, nil
}

// IsBranchMerged checks if the given branch is fully merged into the base branch.
// Returns true if merged, along with the merge commit hash if found.
func (g *GitFlow) IsBranchMerged(branchName, baseBranch string) (bool, *plumbing.Hash, error) {
	// Get the branch reference
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	branchRef, err := g.repo.Reference(branchRefName, true)
	if err == plumbing.ErrReferenceNotFound {
		// Branch doesn't exist - check if it was merged and deleted
		return g.wasBranchMergedAndDeleted(branchName, baseBranch)
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get the base branch reference
	baseRefName := plumbing.NewBranchReferenceName(baseBranch)
	baseRef, err := g.repo.Reference(baseRefName, true)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get base branch reference: %w", err)
	}

	// Check if all commits in the branch are reachable from the base
	branchCommit, err := g.repo.CommitObject(branchRef.Hash())
	if err != nil {
		return false, nil, fmt.Errorf("failed to get branch commit: %w", err)
	}

	baseCommit, err := g.repo.CommitObject(baseRef.Hash())
	if err != nil {
		return false, nil, fmt.Errorf("failed to get base commit: %w", err)
	}

	// Use IsAncestor to check if branch is merged
	isAncestor, err := branchCommit.IsAncestor(baseCommit)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check ancestry: %w", err)
	}

	if isAncestor {
		return true, &baseCommit.Hash, nil
	}

	return false, nil, nil
}

// wasBranchMergedAndDeleted checks if a branch that no longer exists was previously merged.
// This is done by searching commit messages in the base branch for merge commits.
func (g *GitFlow) wasBranchMergedAndDeleted(branchName, baseBranch string) (bool, *plumbing.Hash, error) {
	baseRefName := plumbing.NewBranchReferenceName(baseBranch)
	baseRef, err := g.repo.Reference(baseRefName, true)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get base branch reference: %w", err)
	}

	// Iterate through commits in base branch looking for merge commits
	iter, err := g.repo.Log(&git.LogOptions{From: baseRef.Hash()})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get log: %w", err)
	}
	defer iter.Close()

	// Look for merge commit messages mentioning the branch
	// Typical merge messages: "Merge branch 'branch-name'" or "Merge pull request #123 from branch-name"
	err = iter.ForEach(func(c *object.Commit) error {
		if c.NumParents() >= 2 { // Merge commit
			msg := c.Message
			// Simple check - could be made more sophisticated
			if containsBranchName(msg, branchName) {
				return fmt.Errorf("found") // Use error to break iteration
			}
		}
		return nil
	})

	// If we found a matching merge commit
	if err != nil && err.Error() == "found" {
		hash := baseRef.Hash()
		return true, &hash, nil
	}

	return false, nil, nil
}

// SwitchBranch checks out the specified branch.
func (g *GitFlow) SwitchBranch(branchName string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	branchRefName := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// GetMainBranch attempts to determine the main branch name (main or master).
func (g *GitFlow) GetMainBranch() (string, error) {
	// Try "main" first
	mainRef := plumbing.NewBranchReferenceName("main")
	_, err := g.repo.Reference(mainRef, true)
	if err == nil {
		return "main", nil
	}

	// Try "master"
	masterRef := plumbing.NewBranchReferenceName("master")
	_, err = g.repo.Reference(masterRef, true)
	if err == nil {
		return "master", nil
	}

	return "", fmt.Errorf("could not find main or master branch")
}

// IsWorkingTreeClean returns true if the working tree has no uncommitted changes.
func (g *GitFlow) IsWorkingTreeClean() (bool, error) {
	w, err := g.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return status.IsClean(), nil
}
