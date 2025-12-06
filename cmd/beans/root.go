package beans

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"hmans.dev/beans/internal/bean"
	"hmans.dev/beans/internal/config"
)

var store *bean.Store
var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "beans",
	Short: "A file-based issue tracker for AI-first workflows",
	Long: `Beans is a lightweight issue tracker that stores issues as markdown files
with YAML front matter in a .beans/ directory. Perfect for AI-assisted
development workflows where issues live alongside your code.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip store initialization for init command
		if cmd.Name() == "init" {
			return nil
		}

		root, err := bean.FindRoot()
		if err != nil {
			return fmt.Errorf("no .beans directory found (run 'beans init' to create one)")
		}
		store = bean.NewStore(root)

		cfg, err = config.Load(root)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
