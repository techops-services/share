package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCmd(f *flags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a shared page",
		Long:  "Delete a shared page by its ID. Requires an API key.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			cfg, err := loadConfig(f)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if cfg.APIKey == "" {
				return fmt.Errorf("API key required for deleting shares.\n  Set one with: share init\n  Or pass: --api-key <key>")
			}

			c := newClient(cfg)
			if err := c.Delete(id); err != nil {
				return fmt.Errorf("deleting share %q: %w", id, err)
			}

			fmt.Printf("Deleted %s\n", id)
			return nil
		},
	}
}
