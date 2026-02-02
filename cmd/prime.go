package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/config"
)

//go:embed prompt.tmpl
var agentPromptTemplate string

//go:embed prompt_minimal.tmpl
var agentPromptMinimalTemplate string

// promptData holds all data needed to render the prompt template.
type promptData struct {
	GraphQLSchema string
	Types         []config.TypeConfig
	Statuses      []config.StatusConfig
	Priorities    []config.PriorityConfig
}

var (
	primeMinimal bool
	primeExport  bool
)

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output instructions for AI coding agents",
	Long: `Outputs a prompt that primes AI coding agents on how to use the beans CLI to manage project issues.

Use --minimal for a compact version (~50 tokens) suitable for context-limited scenarios.
Use --export to output the default template for customization.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --export flag
		if primeExport {
			fmt.Print(agentPromptTemplate)
			return nil
		}

		// If no explicit path given, check if a beans project exists by searching
		// upward for a .beans.yml config file
		if beansPath == "" && configPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // Silently exit on error
			}
			configFile, err := config.FindConfig(cwd)
			if err != nil || configFile == "" {
				// No config file found - silently exit
				return nil
			}
		}

		// Select template based on --minimal flag
		templateContent := agentPromptTemplate
		if primeMinimal {
			templateContent = agentPromptMinimalTemplate
		}

		tmpl, err := template.New("prompt").Parse(templateContent)
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
	primeCmd.Flags().BoolVar(&primeMinimal, "minimal", false, "Output compact version for context-limited scenarios")
	primeCmd.Flags().BoolVar(&primeExport, "export", false, "Output default template content (for customization)")
	rootCmd.AddCommand(primeCmd)
}
