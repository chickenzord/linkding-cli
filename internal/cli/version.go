package cli

import (
	"fmt"

	"github.com/chickenzord/linkding-cli/internal/output"
	"github.com/chickenzord/linkding-cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Example: `  linkding version
  linkding version --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := map[string]string{
				"version":    version.Version,
				"git_commit": version.GitCommit,
				"build_date": version.BuildDate,
			}
			if output.IsJSONMode {
				output.JSON(info)
				return nil
			}
			fmt.Printf("linkding %s (commit: %s, built: %s)\n",
				version.Version, version.GitCommit, version.BuildDate)
			return nil
		},
	}
}
