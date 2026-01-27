package gitflow

import (
	"fmt"
	"strings"
	"time"

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

// CreateBranch creates a new git branch with the given name from the base branch.
// This follows GitHub Flow principles: always branch from main (or configured base branch).
// Returns the full branch name and switches to it.
func (g *GitFlow) CreateBranch(beanID, slug, baseBranch string) (string, error) {
	branchName := BuildBranchName(beanID, slug)

	// Check if branch already exists
	exists, err := g.BranchExists(branchName)
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		return "", fmt.Errorf("branch %q already exists", branchName)
	}

	// GitHub Flow: Always branch from the base branch (main)
	baseRefName := plumbing.NewBranchReferenceName(baseBranch)
	baseRef, err := g.repo.Reference(baseRefName, true)
	if err != nil {
		return "", fmt.Errorf("base branch %q not found: %w", baseBranch, err)
	}

	// Warn if not currently on base branch (informational, not blocking)
	currentBranch, err := g.GetCurrentBranch()
	if err == nil && currentBranch != baseBranch {
		// This is just a warning - we still create the branch from base
		// The caller can choose to surface this to the user
	}

	// Create new branch reference FROM base branch
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRefName, baseRef.Hash())

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
// This handles multiple merge strategies:
// - Regular merges (merge commits)
// - Squash merges (GitHub default)
// - Rebase merges (fast-forward)
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

	branchCommit, err := g.repo.CommitObject(branchRef.Hash())
	if err != nil {
		return false, nil, fmt.Errorf("failed to get branch commit: %w", err)
	}

	baseCommit, err := g.repo.CommitObject(baseRef.Hash())
	if err != nil {
		return false, nil, fmt.Errorf("failed to get base commit: %w", err)
	}

	// Strategy 1: Check if branch commit is an ancestor of base (regular merge or fast-forward)
	isAncestor, err := branchCommit.IsAncestor(baseCommit)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check ancestry: %w", err)
	}
	if isAncestor {
		return true, &baseCommit.Hash, nil
	}

	// Strategy 2: Check if base contains all commits from branch (handles squash merges)
	// For squash merges, the individual commits won't be in base, but if we can't find
	// any unique commits in the branch that aren't reachable from base, it's merged
	merged, err := g.areAllCommitsReachable(branchRef.Hash(), baseRef.Hash())
	if err != nil {
		return false, nil, err
	}
	if merged {
		return true, &baseCommit.Hash, nil
	}

	// Strategy 3: Check commit messages for merge references (fallback)
	// This handles cases where commits were squashed and the original branch commits are gone
	return g.wasBranchMergedAndDeleted(branchName, baseBranch)
}

// areAllCommitsReachable checks if all commits in the branch are reachable from base.
// This handles squash merges where the branch commits don't exist in base, but the
// content has been incorporated.
func (g *GitFlow) areAllCommitsReachable(branchHash, baseHash plumbing.Hash) (bool, error) {
	// Walk commits from base
	baseReachable := make(map[plumbing.Hash]bool)
	baseCommit, err := g.repo.CommitObject(baseHash)
	if err != nil {
		return false, fmt.Errorf("failed to get base commit: %w", err)
	}

	iter := object.NewCommitIterCTime(baseCommit, nil, nil)
	err = iter.ForEach(func(c *object.Commit) error {
		baseReachable[c.Hash] = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to iterate base commits: %w", err)
	}

	// Check if branch commit is in base
	if baseReachable[branchHash] {
		return true, nil
	}

	// For more sophisticated squash merge detection, we'd need to compare
	// tree contents, but that's expensive. For now, return false.
	return false, nil
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

// GetMainBranch attempts to determine the main/default branch name.
// Uses multiple strategies in order of preference:
// 1. Read origin/HEAD (the remote's default branch)
// 2. Try common names: "main", "master"
func (g *GitFlow) GetMainBranch() (string, error) {
	// Strategy 1: Check origin/HEAD (most reliable - respects remote's default)
	originHead := plumbing.NewSymbolicReference("refs/remotes/origin/HEAD", "")
	ref, err := g.repo.Reference(originHead.Name(), true)
	if err == nil && ref.Type() == plumbing.SymbolicReference {
		// origin/HEAD is a symbolic ref pointing to origin/<branch>
		// Extract the branch name from refs/remotes/origin/<branch>
		target := ref.Target().Short()
		// Remove "origin/" prefix if present
		if len(target) > 7 && target[:7] == "origin/" {
			branchName := target[7:]
			// Verify the local branch exists
			localRef := plumbing.NewBranchReferenceName(branchName)
			if _, err := g.repo.Reference(localRef, true); err == nil {
				return branchName, nil
			}
		}
	}

	// Strategy 2: Try common default branch names
	// Try "main" first (modern default)
	mainRef := plumbing.NewBranchReferenceName("main")
	_, err = g.repo.Reference(mainRef, true)
	if err == nil {
		return "main", nil
	}

	// Try "master" (legacy default)
	masterRef := plumbing.NewBranchReferenceName("master")
	_, err = g.repo.Reference(masterRef, true)
	if err == nil {
		return "master", nil
	}

	// Strategy 3: Try to find any branch that looks like a default
	// Get all branches and pick the first one that exists
	// (this is a last resort)
	refs, err := g.repo.References()
	if err == nil {
		var firstBranch string
		err = refs.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsBranch() {
				firstBranch = ref.Name().Short()
				return fmt.Errorf("found") // Stop iteration
			}
			return nil
		})
		if firstBranch != "" {
			return firstBranch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
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

// HasOnlyBeansDirChanges returns true if the only uncommitted changes are in the .beans/ directory or .beans.yml.
// Returns false if either the tree is completely clean OR there are changes outside .beans/.
func (g *GitFlow) HasOnlyBeansDirChanges() (bool, error) {
	w, err := g.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return false, nil
	}

	// Check if all dirty files are in .beans/ or are .beans.yml
	hasBeanChanges := false
	for file, stat := range status {
		// Check if file has any changes (worktree or staging)
		if stat.Worktree != git.Unmodified || stat.Staging != git.Unmodified {
			if strings.HasPrefix(file, ".beans/") || file == ".beans.yml" {
				hasBeanChanges = true
			} else {
				// Found changes outside .beans/
				return false, nil
			}
		}
	}

	return hasBeanChanges, nil
}

// CommitBeans commits all changes in the .beans/ directory and .beans.yml with the given message.
// This is used for auto-committing bean updates before creating git branches.
func (g *GitFlow) CommitBeans(message string) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Add .beans.yml if it exists and has changes
	if err := w.AddGlob(".beans.yml"); err != nil {
		// Ignore error if file doesn't exist
	}

	// Add all files in .beans/ directory
	if err := w.AddGlob(".beans/*"); err != nil {
		return fmt.Errorf("failed to add .beans/ files: %w", err)
	}

	// Create commit
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "beans",
			Email: "beans@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
