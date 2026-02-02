package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
	"github.com/hmans/beans/internal/output"
)

var (
	initJSON       bool
	initClaudeHooks bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a beans project",
	Long: `Creates a .beans directory and .beans.yml config file in the current directory.

Use --claude-hooks to also configure Claude Code hooks for automatic beans prime injection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectDir string
		var beansDir string
		var dirName string

		if beansPath != "" {
			// Use explicit path for beans directory
			beansDir = beansPath
			projectDir = filepath.Dir(beansDir)
			dirName = filepath.Base(projectDir)
			// Create the directory using Core.Init to set up .gitignore
			core := beancore.New(beansDir, nil)
			if err := core.Init(); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			// Use current working directory
			dir, err := os.Getwd()
			if err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return err
			}

			if err := beancore.Init(dir); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to initialize: %w", err)
			}

			projectDir = dir
			beansDir = filepath.Join(dir, ".beans")
			dirName = filepath.Base(dir)
		}

		// Create default config file with directory name as prefix
		// Config is saved at project root (not inside .beans/)
		defaultCfg := config.DefaultWithPrefix(dirName + "-")
		defaultCfg.SetConfigDir(projectDir)
		if err := defaultCfg.Save(projectDir); err != nil {
			if initJSON {
				return output.Error(output.ErrFileError, err.Error())
			}
			return fmt.Errorf("failed to create config: %w", err)
		}

		// Configure Claude Code hooks if requested
		if initClaudeHooks {
			if err := configureClaudeHooks(projectDir); err != nil {
				if initJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to configure Claude hooks: %w", err)
			}
		}

		if initJSON {
			return output.SuccessInit(beansDir)
		}

		fmt.Println("Initialized beans project")
		if initClaudeHooks {
			fmt.Println("Configured Claude Code hooks in .claude/settings.json")
		}
		return nil
	},
}

// configureClaudeHooks creates or updates .claude/settings.json with beans prime hooks
func configureClaudeHooks(projectDir string) error {
	claudeDir := filepath.Join(projectDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Create .claude directory if it doesn't exist
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Load existing settings or create new
	settings := make(map[string]any)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	}

	// Define the beans prime hook
	beansPrimeHook := map[string]any{
		"hooks": []map[string]any{
			{
				"type":    "command",
				"command": "beans prime",
			},
		},
	}

	// Get or create hooks section
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = make(map[string]any)
		settings["hooks"] = hooks
	}

	// Add SessionStart hook if not already configured for beans
	if !hasBeansHook(hooks, "SessionStart") {
		sessionStart, _ := hooks["SessionStart"].([]any)
		sessionStart = append(sessionStart, beansPrimeHook)
		hooks["SessionStart"] = sessionStart
	}

	// Add PreCompact hook if not already configured for beans
	if !hasBeansHook(hooks, "PreCompact") {
		preCompact, _ := hooks["PreCompact"].([]any)
		preCompact = append(preCompact, beansPrimeHook)
		hooks["PreCompact"] = preCompact
	}

	// Write settings back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// hasBeansHook checks if a hook type already has a beans prime hook configured
func hasBeansHook(hooks map[string]any, hookType string) bool {
	hookList, ok := hooks[hookType].([]any)
	if !ok {
		return false
	}

	for _, hook := range hookList {
		hookMap, ok := hook.(map[string]any)
		if !ok {
			continue
		}

		// Check nested hooks array
		innerHooks, ok := hookMap["hooks"].([]any)
		if !ok {
			continue
		}

		for _, inner := range innerHooks {
			innerMap, ok := inner.(map[string]any)
			if !ok {
				continue
			}

			if cmd, ok := innerMap["command"].(string); ok {
				if cmd == "beans prime" {
					return true
				}
			}
		}
	}

	return false
}

func init() {
	initCmd.Flags().BoolVar(&initJSON, "json", false, "Output as JSON")
	initCmd.Flags().BoolVar(&initClaudeHooks, "claude-hooks", false, "Configure Claude Code hooks for beans prime")
	rootCmd.AddCommand(initCmd)
}
