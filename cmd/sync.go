package cmd

import (
	"context"
	"fmt"

	"github.com/hmans/beans/internal/graph"
	"github.com/hmans/beans/internal/output"
	"github.com/spf13/cobra"
)

var (
	syncApply bool
	syncJSON  bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize bean status with git branch lifecycle",
	Long: `Checks all beans with git branches and updates their status based on branch state:
- Merged branches → bean status becomes 'completed'
- Deleted branches (not merged) → bean status becomes 'scrapped'

By default, shows a preview of changes without applying them.
Use --apply to actually update the beans.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if git integration is enabled
		if !core.IsGitFlowEnabled() {
			// Try to enable git integration
			if err := core.EnableGitFlow("."); err != nil {
				return cmdError(syncJSON, output.ErrGit, "git integration not available: %v", err)
			}
		}

		ctx := context.Background()
		resolver := &graph.Resolver{Core: core}

		// Perform the sync
		updatedBeans, err := resolver.Mutation().SyncGitBranches(ctx)
		if err != nil {
			return cmdError(syncJSON, output.ErrGit, "sync failed: %v", err)
		}

		if syncJSON {
			return output.JSON(output.Response{
				Success: true,
				Beans:   updatedBeans,
				Count:   len(updatedBeans),
				Message: "Git branches synced",
			})
		}

		// Human-readable output
		if len(updatedBeans) == 0 {
			fmt.Println("No beans to sync")
			return nil
		}

		if !syncApply {
			// Preview mode
			fmt.Printf("Found %d bean(s) with git integration:\n\n", len(updatedBeans))
			for _, b := range updatedBeans {
				statusEmoji := "✓"
				var reason string
				switch b.Status {
				case "completed":
					reason = "Branch merged → would mark as completed"
				case "scrapped":
					reason = "Branch deleted → would mark as scrapped"
				default:
					reason = "Would update status"
				}
				fmt.Printf("  %s %s  %-30s  %s\n", statusEmoji, b.ID, b.Title, reason)
			}
			fmt.Println("\nRun with --apply to update beans")
			return nil
		}

		// Apply mode - beans are already updated by mutation
		fmt.Printf("Synced %d bean(s):\n\n", len(updatedBeans))
		for _, b := range updatedBeans {
			statusEmoji := "✓"
			var reason string
			switch b.Status {
			case "completed":
				if b.GitBranch != "" {
					reason = fmt.Sprintf("Marked as completed (branch %s merged)", b.GitBranch)
				} else {
					reason = "Marked as completed"
				}
			case "scrapped":
				if b.GitBranch != "" {
					reason = fmt.Sprintf("Marked as scrapped (branch %s deleted)", b.GitBranch)
				} else {
					reason = "Marked as scrapped"
				}
			default:
				reason = "Status updated"
			}
			fmt.Printf("  %s %s  %s\n", statusEmoji, b.ID, reason)
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncApply, "apply", false, "Apply changes (default: dry-run preview)")
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(syncCmd)
}
