package cli

import (
	"fmt"
	"os"

	"github.com/chickenzord/linkding-cli/internal/client"
	"github.com/chickenzord/linkding-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
		Long:  "Commands to list, get, and create tags.",
	}
	cmd.AddCommand(
		newTagListCmd(),
		newTagGetCmd(),
		newTagCreateCmd(),
	)
	return cmd
}

func newTagListCmd() *cobra.Command {
	var query string
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		Example: `  linkding tag list
  linkding tag list --json
  linkding tag list -q`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			result, err := c.ListTags(client.TagListOptions{Query: query, Limit: limit, Offset: offset})
			if err != nil {
				handleError(err, "list_failed", nil, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(result)
				return nil
			}

			if output.IsQuietMode {
				for _, t := range result.Results {
					fmt.Println(t.Name)
				}
				return nil
			}

			tw := output.NewTabWriter()
			fmt.Fprintf(tw, "ID\tNAME\n")
			fmt.Fprintf(tw, "--\t----\n")
			for _, t := range result.Results {
				fmt.Fprintf(tw, "%d\t%s\n", t.ID, t.Name)
			}
			tw.Flush()
			output.Infof("\nTotal: %d\n", result.Count)
			return nil
		},
	}

	cmd.Flags().StringVarP(&query, "search", "s", "", "Filter tags by name")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max results to return")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	return cmd
}

func newTagGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a tag by ID",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding tag get 5
  linkding tag get 5 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric tag ID", false)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "", false)
				os.Exit(ExitUsage)
			}

			t, err := c.GetTag(id)
			if err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("tag %d not found", id),
						map[string]int{"id": id}, "Run 'linkding tag list' to see available tags", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "get_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(t)
				return nil
			}

			tw := output.NewTabWriter()
			fmt.Fprintf(tw, "ID:\t%d\n", t.ID)
			fmt.Fprintf(tw, "Name:\t%s\n", t.Name)
			tw.Flush()
			return nil
		},
	}
}

func newTagCreateCmd() *cobra.Command {
	var name string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tag",
		Example: `  linkding tag create --name golang
  linkding tag create --name golang --json
  linkding tag create --name golang --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				output.WriteError("invalid_input", "--name is required", nil, "Provide a tag name with --name", false)
				os.Exit(ExitUsage)
			}

			// Validate: reject control chars and shell-special chars in tag names
			for _, r := range name {
				if r < 0x20 || r == '`' || r == '$' || r == '\\' {
					output.WriteError("invalid_input", fmt.Sprintf("tag name %q contains invalid characters", name),
						map[string]string{"name": name}, "", false)
					os.Exit(ExitUsage)
				}
			}

			if dryRun {
				output.Infof("[dry-run] Would create tag: %q\n", name)
				return nil
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "", false)
				os.Exit(ExitUsage)
			}

			t, err := c.CreateTag(client.TagInput{Name: name})
			if err != nil {
				handleError(err, "create_failed", map[string]string{"name": name}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(t)
				return nil
			}

			output.Infof("Created tag %q (ID: %d)\n", t.Name, t.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Tag name (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without making changes")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
