package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/hmans/beans/internal/output"
)

var (
	forceDelete bool
	deleteJSON  bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a bean",
	Long: `Deletes a bean after confirmation (use -f to skip confirmation).

If other beans link to this bean, you will be warned and those references
will be removed after confirmation. Use -f to skip all warnings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := core.Get(args[0])
		if err != nil {
			if deleteJSON {
				return output.Error(output.ErrNotFound, err.Error())
			}
			return fmt.Errorf("failed to find bean: %w", err)
		}

		// Check for incoming links
		incomingLinks := core.FindIncomingLinks(b.ID)
		hasIncoming := len(incomingLinks) > 0

		// JSON implies force (no prompts for machines)
		if !forceDelete && !deleteJSON {
			// Warn about incoming links
			if hasIncoming {
				fmt.Printf("Warning: %d bean(s) link to '%s':\n", len(incomingLinks), b.Title)
				for _, link := range incomingLinks {
					fmt.Printf("  - %s (%s) via %s\n", link.FromBean.ID, link.FromBean.Title, link.LinkType)
				}
				fmt.Print("Delete anyway and remove references? [y/N] ")
			} else {
				fmt.Printf("Delete '%s' (%s)? [y/N] ", b.Title, b.Path)
			}

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		// Remove incoming links before deletion
		if hasIncoming {
			removed, err := core.RemoveLinksTo(b.ID)
			if err != nil {
				if deleteJSON {
					return output.Error(output.ErrFileError, fmt.Sprintf("failed to remove references: %s", err))
				}
				return fmt.Errorf("failed to remove references: %w", err)
			}
			if !deleteJSON {
				fmt.Printf("Removed %d reference(s)\n", removed)
			}
		}

		if err := core.Delete(args[0]); err != nil {
			if deleteJSON {
				return output.Error(output.ErrFileError, err.Error())
			}
			return fmt.Errorf("failed to delete bean: %w", err)
		}

		if deleteJSON {
			return output.Success(b, "Bean deleted")
		}

		fmt.Printf("Deleted %s\n", b.Path)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation and warnings")
	deleteCmd.Flags().BoolVar(&deleteJSON, "json", false, "Output as JSON (implies --force)")
	rootCmd.AddCommand(deleteCmd)
}
