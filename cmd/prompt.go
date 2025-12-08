package cmd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"hmans.dev/beans/internal/beancore"
	"hmans.dev/beans/internal/config"
)

// Note: config import is still needed for DefaultTypes and DefaultStatuses

//go:embed prompt.md
var agentPrompt string

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Output instructions for AI coding agents",
	Long:  `Outputs a prompt that instructs AI coding agents on how to use the beans CLI to manage project issues.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find beans directory; silently exit if none exists
		if beansPath == "" {
			_, err := beancore.FindRoot()
			if err != nil {
				// No .beans directory found - silently exit
				return nil
			}
		}

		fmt.Print(agentPrompt)

		// Append dynamic sections (types and statuses are hardcoded)
		var sb strings.Builder

		// Issue types section (hardcoded types)
		sb.WriteString("\n## Issue Types\n\n")
		sb.WriteString("This project has the following issue types configured. Always specify a type with `-t` when creating beans:\n\n")
		for _, t := range config.DefaultTypes {
			if t.Description != "" {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", t.Name, t.Description))
			} else {
				sb.WriteString(fmt.Sprintf("- **%s**\n", t.Name))
			}
		}

		// Statuses section (hardcoded statuses)
		sb.WriteString("\n## Statuses\n\n")
		sb.WriteString("This project has the following statuses configured:\n\n")
		for _, s := range config.DefaultStatuses {
			if s.Description != "" {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", s.Name, s.Description))
			} else {
				sb.WriteString(fmt.Sprintf("- **%s**\n", s.Name))
			}
		}

		sb.WriteString("\n")
		fmt.Print(sb.String())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
