package gitflow

// This file contains git synchronization utilities.
// The actual sync logic that modifies beans is in beancore to avoid circular imports.

// BranchStatus represents the status of a git branch.
type BranchStatus int

const (
	BranchStatusUnknown BranchStatus = iota
	BranchStatusActive   // Branch exists and has unmerged commits
	BranchStatusMerged   // Branch is fully merged into base
	BranchStatusDeleted  // Branch doesn't exist (deleted without merge)
)

// GetBranchStatus determines the status of a branch.
func (g *GitFlow) GetBranchStatus(branchName, baseBranch string) (BranchStatus, error) {
	// Check if branch exists
	exists, err := g.BranchExists(branchName)
	if err != nil {
		return BranchStatusUnknown, err
	}

	if !exists {
		// Branch doesn't exist - check if it was merged and deleted
		merged, _, err := g.IsBranchMerged(branchName, baseBranch)
		if err != nil {
			return BranchStatusUnknown, err
		}

		if merged {
			return BranchStatusMerged, nil
		}
		return BranchStatusDeleted, nil
	}

	// Branch exists - check if it's merged
	merged, _, err := g.IsBranchMerged(branchName, baseBranch)
	if err != nil {
		return BranchStatusUnknown, err
	}

	if merged {
		return BranchStatusMerged, nil
	}

	return BranchStatusActive, nil
}
