package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chickenzord/linkding-cli/internal/client"
	"github.com/chickenzord/linkding-cli/internal/output"
	"github.com/spf13/cobra"
)

func newBookmarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookmark",
		Short: "Manage bookmarks",
		Long:  "Commands to list, create, update, delete, and manage bookmarks.",
	}

	cmd.AddCommand(
		newBookmarkListCmd(),
		newBookmarkGetCmd(),
		newBookmarkCreateCmd(),
		newBookmarkUpdateCmd(),
		newBookmarkDeleteCmd(),
		newBookmarkCheckCmd(),
		newBookmarkArchiveCmd(),
		newBookmarkUnarchiveCmd(),
	)
	return cmd
}

// ---- list ----

func newBookmarkListCmd() *cobra.Command {
	var search string
	var archived, unread, shared bool
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bookmarks",
		Example: `  # List all bookmarks
  linkding bookmark list

  # Search for bookmarks
  linkding bookmark list --search "golang"

  # List archived bookmarks as JSON
  linkding bookmark list --archived --json

  # Pipe-friendly: print just URLs
  linkding bookmark list -q`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			result, err := c.ListBookmarks(client.BookmarkListOptions{
				Query:    search,
				Archived: archived,
				Unread:   unread,
				Shared:   shared,
				Limit:    limit,
				Offset:   offset,
			})
			if err != nil {
				handleError(err, "list_failed", map[string]interface{}{
					"search": search, "archived": archived,
				}, "Check your credentials and URL", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(result)
				return nil
			}

			if output.IsQuietMode {
				for _, b := range result.Results {
					fmt.Println(b.URL)
				}
				return nil
			}

			tw := output.NewTabWriter()
			fmt.Fprintf(tw, "ID\tURL\tTITLE\tTAGS\tARCHIVED\n")
			fmt.Fprintf(tw, "--\t---\t-----\t----\t--------\n")
			for _, b := range result.Results {
				title := b.Title
				if title == "" {
					title = b.WebsiteTitle
				}
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				u := b.URL
				if len(u) > 50 {
					u = u[:47] + "..."
				}
				fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%v\n",
					b.ID, u, title, strings.Join(b.TagNames, ","), b.IsArchived)
			}
			tw.Flush()
			output.Infof("\nTotal: %d\n", result.Count)
			return nil
		},
	}

	cmd.Flags().StringVarP(&search, "search", "s", "", "Search query")
	cmd.Flags().BoolVar(&archived, "archived", false, "List archived bookmarks")
	cmd.Flags().BoolVar(&unread, "unread", false, "List unread bookmarks")
	cmd.Flags().BoolVar(&shared, "shared", false, "List shared bookmarks")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max results to return")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")
	return cmd
}

// ---- get ----

func newBookmarkGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a bookmark by ID",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding bookmark get 42
  linkding bookmark get 42 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q: must be a positive integer", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric bookmark ID", false)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			b, err := c.GetBookmark(id)
			if err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("bookmark %d not found", id),
						map[string]int{"id": id}, "Run 'linkding bookmark list' to see available bookmarks", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "get_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(b)
				return nil
			}

			printBookmarkDetail(b)
			return nil
		},
	}
}

// ---- create ----

func newBookmarkCreateCmd() *cobra.Command {
	var bURL, title, description, notes string
	var tagNames []string
	var isArchived, unread, shared bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bookmark",
		Example: `  linkding bookmark create --url https://example.com --title "Example" --tags go,tools
  linkding bookmark create --url https://example.com --json
  linkding bookmark create --url https://example.com --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if bURL == "" {
				output.WriteError("invalid_input", "--url is required", nil, "Provide a URL with --url", false)
				os.Exit(ExitUsage)
			}
			if err := validateURL(bURL); err != nil {
				output.WriteError("invalid_input", err.Error(), map[string]string{"url": bURL}, "", false)
				os.Exit(ExitUsage)
			}

			input := client.BookmarkInput{
				URL:         bURL,
				Title:       title,
				Description: description,
				Notes:       notes,
				TagNames:    tagNames,
				IsArchived:  isArchived,
				Unread:      unread,
				Shared:      shared,
			}

			if dryRun {
				output.Infof("[dry-run] Would create bookmark:\n")
				output.JSON(input)
				return nil
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			b, err := c.CreateBookmark(input)
			if err != nil {
				handleError(err, "create_failed", map[string]string{"url": bURL}, "Check that the URL is valid and the token has write access", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(b)
				return nil
			}

			output.Infof("Created bookmark (ID: %d)\n", b.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&bURL, "url", "", "Bookmark URL (required)")
	cmd.Flags().StringVar(&title, "title", "", "Title")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVar(&notes, "notes", "", "Notes (Markdown)")
	cmd.Flags().StringSliceVar(&tagNames, "tags", nil, "Comma-separated tag names")
	cmd.Flags().BoolVar(&isArchived, "archived", false, "Create as archived")
	cmd.Flags().BoolVar(&unread, "unread", false, "Mark as unread")
	cmd.Flags().BoolVar(&shared, "shared", false, "Share publicly")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created without making changes")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

// ---- update ----

func newBookmarkUpdateCmd() *cobra.Command {
	var bURL, title, description, notes string
	var tagNames []string
	var isArchived, unread, shared bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing bookmark (PATCH)",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding bookmark update 42 --title "New Title"
  linkding bookmark update 42 --tags go,programming --json
  linkding bookmark update 42 --url https://new.example.com --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric bookmark ID", false)
				os.Exit(ExitUsage)
			}

			if bURL != "" {
				if err := validateURL(bURL); err != nil {
					output.WriteError("invalid_input", err.Error(), map[string]string{"url": bURL}, "", false)
					os.Exit(ExitUsage)
				}
			}

			input := client.BookmarkInput{
				URL:         bURL,
				Title:       title,
				Description: description,
				Notes:       notes,
				TagNames:    tagNames,
				IsArchived:  isArchived,
				Unread:      unread,
				Shared:      shared,
			}

			if dryRun {
				output.Infof("[dry-run] Would update bookmark %d with:\n", id)
				output.JSON(input)
				return nil
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			b, err := c.UpdateBookmark(id, input)
			if err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("bookmark %d not found", id),
						map[string]int{"id": id}, "Run 'linkding bookmark list' to see available bookmarks", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "update_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(b)
				return nil
			}

			output.Infof("Updated bookmark %d\n", b.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&bURL, "url", "", "New URL")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&notes, "notes", "", "New notes (Markdown)")
	cmd.Flags().StringSliceVar(&tagNames, "tags", nil, "Comma-separated tag names (replaces existing)")
	cmd.Flags().BoolVar(&isArchived, "archived", false, "Mark as archived")
	cmd.Flags().BoolVar(&unread, "unread", false, "Mark as unread")
	cmd.Flags().BoolVar(&shared, "shared", false, "Share publicly")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	return cmd
}

// ---- delete ----

func newBookmarkDeleteCmd() *cobra.Command {
	var yes bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a bookmark",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding bookmark delete 42 --yes
  linkding bookmark delete 42 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric bookmark ID", false)
				os.Exit(ExitUsage)
			}

			if dryRun {
				output.Infof("[dry-run] Would delete bookmark %d\n", id)
				return nil
			}

			// Require --yes or non-TTY context
			if !yes && isTerminal() {
				output.Errf("Error: this is a destructive action. Pass --yes to confirm deletion of bookmark %d\n", id)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			if err := c.DeleteBookmark(id); err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("bookmark %d not found", id),
						map[string]int{"id": id}, "Run 'linkding bookmark list' to see available bookmarks", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "delete_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(map[string]interface{}{"deleted": true, "id": id})
				return nil
			}
			output.Infof("Deleted bookmark %d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without making changes")
	return cmd
}

// ---- check ----

func newBookmarkCheckCmd() *cobra.Command {
	var bURL string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if a URL is already bookmarked and fetch metadata",
		Example: `  linkding bookmark check --url https://example.com
  linkding bookmark check --url https://example.com --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if bURL == "" {
				output.WriteError("invalid_input", "--url is required", nil, "Provide a URL with --url", false)
				os.Exit(ExitUsage)
			}
			if err := validateURL(bURL); err != nil {
				output.WriteError("invalid_input", err.Error(), map[string]string{"url": bURL}, "", false)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "Set --url and --token or LINKDING_URL and LINKDING_TOKEN", false)
				os.Exit(ExitUsage)
			}

			result, err := c.CheckBookmark(bURL)
			if err != nil {
				handleError(err, "check_failed", map[string]string{"url": bURL}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(result)
				return nil
			}

			if result.Bookmark != nil {
				output.Infof("URL is already bookmarked (ID: %d)\n", result.Bookmark.ID)
			} else {
				output.Infof("URL is not bookmarked\n")
			}
			output.Infof("Title: %s\n", result.Metadata.Title)
			output.Infof("Description: %s\n", result.Metadata.Description)
			if len(result.AutoTags) > 0 {
				output.Infof("Suggested tags: %s\n", strings.Join(result.AutoTags, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&bURL, "url", "", "URL to check (required)")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

// ---- archive / unarchive ----

func newBookmarkArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a bookmark",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding bookmark archive 42
  linkding bookmark archive 42 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric bookmark ID", false)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "", false)
				os.Exit(ExitUsage)
			}

			if err := c.ArchiveBookmark(id); err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("bookmark %d not found", id),
						map[string]int{"id": id}, "Run 'linkding bookmark list' to see available bookmarks", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "archive_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(map[string]interface{}{"archived": true, "id": id})
				return nil
			}
			output.Infof("Archived bookmark %d\n", id)
			return nil
		},
	}
}

func newBookmarkUnarchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <id>",
		Short: "Unarchive a bookmark",
		Args:  cobra.ExactArgs(1),
		Example: `  linkding bookmark unarchive 42
  linkding bookmark unarchive 42 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				output.WriteError("invalid_input", fmt.Sprintf("invalid ID %q", args[0]),
					map[string]string{"id": args[0]}, "Provide a numeric bookmark ID", false)
				os.Exit(ExitUsage)
			}

			c, err := newClient(cmd)
			if err != nil {
				output.WriteError("config_error", err.Error(), nil, "", false)
				os.Exit(ExitUsage)
			}

			if err := c.UnarchiveBookmark(id); err != nil {
				if isAPINotFound(err) {
					output.WriteError("not_found", fmt.Sprintf("bookmark %d not found", id),
						map[string]int{"id": id}, "Run 'linkding bookmark list' to see available bookmarks", false)
					os.Exit(ExitNotFound)
				}
				handleError(err, "unarchive_failed", map[string]int{"id": id}, "", false)
				os.Exit(exitCode(err))
			}

			if output.IsJSONMode {
				output.JSON(map[string]interface{}{"archived": false, "id": id})
				return nil
			}
			output.Infof("Unarchived bookmark %d\n", id)
			return nil
		},
	}
}

// ---- helpers ----

func parseID(s string) (int, error) {
	// Reject any non-numeric or suspicious characters
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("ID must be a positive integer")
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("ID must be a positive integer")
	}
	return n, nil
}

func validateURL(u string) error {
	if strings.Contains(u, "..") {
		return fmt.Errorf("invalid URL: must not contain '..'")
	}
	// Reject control characters
	for _, r := range u {
		if r < 0x20 {
			return fmt.Errorf("invalid URL: must not contain control characters")
		}
	}
	return nil
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func printBookmarkDetail(b *client.Bookmark) {
	title := b.Title
	if title == "" {
		title = b.WebsiteTitle
	}
	tw := output.NewTabWriter()
	fmt.Fprintf(tw, "ID:\t%d\n", b.ID)
	fmt.Fprintf(tw, "URL:\t%s\n", b.URL)
	fmt.Fprintf(tw, "Title:\t%s\n", title)
	fmt.Fprintf(tw, "Description:\t%s\n", b.Description)
	fmt.Fprintf(tw, "Tags:\t%s\n", strings.Join(b.TagNames, ", "))
	fmt.Fprintf(tw, "Archived:\t%v\n", b.IsArchived)
	fmt.Fprintf(tw, "Unread:\t%v\n", b.Unread)
	fmt.Fprintf(tw, "Shared:\t%v\n", b.Shared)
	fmt.Fprintf(tw, "Added:\t%s\n", b.DateAdded)
	fmt.Fprintf(tw, "Modified:\t%s\n", b.DateModified)
	tw.Flush()
}
