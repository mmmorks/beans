package cmd

import (
	"fmt"
	"os"

	"github.com/hmans/beans/internal/config"
	"github.com/spf13/cobra"
)

const onboardTemplate = `beans Onboarding

Add this minimal snippet to AGENTS.md (or create it):

--- BEGIN AGENTS.MD CONTENT ---
## Issue Tracking

This project uses **beans** for issue tracking.
Run ` + "`beans prime`" + ` for workflow context, or install hooks for auto-injection.

**Quick reference:**
- ` + "`beans list --ready`" + ` - Find unblocked work
- ` + "`beans create \"Title\" -t task -s in-progress`" + ` - Create and start issue
- ` + "`beans update <id> -s completed`" + ` - Complete work
- ` + "`beans sync --apply`" + ` - Sync git branches with bean status

For full workflow details: ` + "`beans prime`" + `
--- END AGENTS.MD CONTENT ---

For Claude Code users:

  Option 1 (automatic): Run ` + "`beans init --claude-hooks`" + `

  Option 2 (manual): Add to .claude/settings.json:
  {
    "hooks": {
      "SessionStart": [{"hooks": [{"type": "command", "command": "beans prime"}]}],
      "PreCompact": [{"hooks": [{"type": "command", "command": "beans prime"}]}]
    }
  }

How it works:
   • beans prime provides dynamic workflow context
   • SessionStart hook auto-injects at session start
   • PreCompact hook re-injects after context compaction
   • AGENTS.md only needs this minimal pointer
`

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Display minimal snippet for AGENTS.md",
	Long:  `Outputs a minimal snippet to add to AGENTS.md that points to 'beans prime' for full workflow context.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if beans is initialized in this directory
		if beansPath == "" && configPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			configFile, err := config.FindConfig(cwd)
			if err != nil || configFile == "" {
				return fmt.Errorf("beans not initialized in this directory (no .beans.yml found)")
			}
		}

		fmt.Print(onboardTemplate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}
