package cmd

import (
	_ "embed"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/config"
)

//go:embed prompt.tmpl
var agentPromptTemplate string

// promptData holds all data needed to render the prompt template.
type promptData struct {
	GraphQLSchema string
	Types         []config.TypeConfig
	Statuses      []config.StatusConfig
	Priorities    []config.PriorityConfig
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Output instructions for AI coding agents",
	Long:  `Outputs a prompt that instructs AI coding agents on how to use the beans CLI to manage project issues.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no explicit path given, check if a beans project exists
		if beansPath == "" && configPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // Silently exit on error
			}
			cfg, err := config.LoadFromDirectory(cwd)
			if err != nil {
				return nil // Silently exit on error
			}
			// Check if the beans directory exists
			beansDir := cfg.ResolveBeansPath()
			if _, err := os.Stat(beansDir); os.IsNotExist(err) {
				// No beans directory found - silently exit
				return nil
			}
		}

		tmpl, err := template.New("prompt").Parse(agentPromptTemplate)
		if err != nil {
			return err
		}

		data := promptData{
			GraphQLSchema: GetGraphQLSchema(),
			Types:         config.DefaultTypes,
			Statuses:      config.DefaultStatuses,
			Priorities:    config.DefaultPriorities,
		}

		return tmpl.Execute(os.Stdout, data)
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
