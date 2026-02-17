package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("share %s\n", buildVersion)
			fmt.Printf("  commit: %s\n", buildCommit)
			fmt.Printf("  built:  %s\n", buildDate)
		},
	}
}
