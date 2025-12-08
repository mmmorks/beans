package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/config"
	"hmans.dev/beans/internal/output"
	"hmans.dev/beans/internal/ui"
)

var (
	createStatus   string
	createType     string
	createPriority string
	createBody     string
	createBodyFile string
	createTag      []string
	createLink     []string
	createJSON     bool
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new bean",
	Long:  `Creates a new bean (issue) with a generated ID and optional title.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		status := createStatus

		// Validate status if provided
		if status != "" && !cfg.IsValidStatus(status) {
			if createJSON {
				return output.Error(output.ErrInvalidStatus, fmt.Sprintf("invalid status: %s (must be %s)", status, cfg.StatusList()))
			}
			return fmt.Errorf("invalid status: %s (must be %s)", status, cfg.StatusList())
		}
		if status == "" {
			status = cfg.GetDefaultStatus()
		}

		// Validate type if provided
		if createType != "" && !cfg.IsValidType(createType) {
			if createJSON {
				return output.Error(output.ErrValidation, fmt.Sprintf("invalid type: %s (must be %s)", createType, cfg.TypeList()))
			}
			return fmt.Errorf("invalid type: %s (must be %s)", createType, cfg.TypeList())
		}
		if createType == "" {
			createType = cfg.GetDefaultType()
		}

		// Validate priority if provided
		if createPriority != "" && !cfg.IsValidPriority(createPriority) {
			if createJSON {
				return output.Error(output.ErrValidation, fmt.Sprintf("invalid priority: %s (must be %s)", createPriority, cfg.PriorityList()))
			}
			return fmt.Errorf("invalid priority: %s (must be %s)", createPriority, cfg.PriorityList())
		}

		// Determine body content
		body, err := resolveContent(createBody, createBodyFile)
		if err != nil {
			if createJSON {
				return output.Error(output.ErrFileError, err.Error())
			}
			return err
		}

		if title == "" {
			title = "Untitled"
		}

		b := &bean.Bean{
			Slug:     bean.Slugify(title),
			Title:    title,
			Status:   status,
			Type:     createType,
			Priority: createPriority,
			Body:     body,
		}

		// Add tags if provided
		if err := applyTags(b, createTag); err != nil {
			if createJSON {
				return output.Error(output.ErrValidation, err.Error())
			}
			return err
		}

		// Add links if provided
		warnings, err := applyLinks(b, createLink)
		if err != nil {
			if createJSON {
				return output.Error(output.ErrValidation, err.Error())
			}
			return err
		}

		if err := core.Create(b); err != nil {
			if createJSON {
				return output.Error(output.ErrFileError, err.Error())
			}
			return fmt.Errorf("failed to create bean: %w", err)
		}

		// Output result
		if createJSON {
			if len(warnings) > 0 {
				return output.SuccessWithWarnings(b, "Bean created", warnings)
			}
			return output.Success(b, "Bean created")
		}

		// Print warnings in text mode
		for _, w := range warnings {
			fmt.Println(ui.Warning.Render("Warning: ") + w)
		}

		fmt.Println(ui.Success.Render("Created ") + ui.ID.Render(b.ID) + " " + ui.Muted.Render(b.Path))

		return nil
	},
}

func init() {
	// Build help text with allowed values from hardcoded config
	statusNames := make([]string, len(config.DefaultStatuses))
	for i, s := range config.DefaultStatuses {
		statusNames[i] = s.Name
	}
	typeNames := make([]string, len(config.DefaultTypes))
	for i, t := range config.DefaultTypes {
		typeNames[i] = t.Name
	}
	priorityNames := make([]string, len(config.DefaultPriorities))
	for i, p := range config.DefaultPriorities {
		priorityNames[i] = p.Name
	}

	createCmd.Flags().StringVarP(&createStatus, "status", "s", "", "Initial status ("+strings.Join(statusNames, ", ")+")")
	createCmd.Flags().StringVarP(&createType, "type", "t", "", "Bean type ("+strings.Join(typeNames, ", ")+")")
	createCmd.Flags().StringVarP(&createPriority, "priority", "p", "", "Priority level ("+strings.Join(priorityNames, ", ")+")")
	createCmd.Flags().StringVarP(&createBody, "body", "d", "", "Body content (use '-' to read from stdin)")
	createCmd.Flags().StringVar(&createBodyFile, "body-file", "", "Read body from file")
	createCmd.Flags().StringArrayVar(&createTag, "tag", nil, "Add tag (can be repeated)")
	createCmd.Flags().StringArrayVar(&createLink, "link", nil, "Add relationship (format: type:id, can be repeated)")
	createCmd.Flags().BoolVar(&createJSON, "json", false, "Output as JSON")
	createCmd.MarkFlagsMutuallyExclusive("body", "body-file")
	rootCmd.AddCommand(createCmd)
}
