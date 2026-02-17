package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/techops-services/share/internal/display"
)

func newListCmd(f *flags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your shared pages",
		Long:  "List all active shared pages. Requires an API key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(f)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if cfg.APIKey == "" {
				return fmt.Errorf("API key required for listing shares.\n  Set one with: share init\n  Or pass: --api-key <key>")
			}

			c := newClient(cfg)
			resp, err := c.List(limit, "")
			if err != nil {
				return fmt.Errorf("listing shares: %w", err)
			}

			display.PrintShareTable(os.Stdout, resp.Shares)

			if f.verbose && resp.Total > 0 {
				fmt.Fprintf(os.Stderr, "\nTotal: %d shares\n", resp.Total)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of shares to list")

	return cmd
}
