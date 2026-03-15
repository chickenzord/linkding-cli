package cli

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/chickenzord/linkding-cli/internal/client"
	"github.com/chickenzord/linkding-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagURL   string
	flagToken string

	// Exit codes
	ExitOK         = 0
	ExitError      = 1
	ExitUsage      = 2
	ExitNotFound   = 3
	ExitForbidden  = 4
	ExitConflict   = 5
)

func newClient(cmd *cobra.Command) (*client.Client, error) {
	u := flagURL
	if u == "" {
		u = os.Getenv("LINKDING_URL")
	}
	if u == "" {
		return nil, fmt.Errorf("linkding URL is required: set --url or LINKDING_URL")
	}

	// Validate URL
	parsed, err := url.Parse(u)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid URL %q: must be http or https", u)
	}
	// Reject path traversal attempts
	if strings.Contains(u, "..") {
		return nil, fmt.Errorf("invalid URL: must not contain '..'")
	}

	tok := flagToken
	if tok == "" {
		tok = os.Getenv("LINKDING_TOKEN")
	}
	if tok == "" {
		return nil, fmt.Errorf("linkding API token is required: set --token or LINKDING_TOKEN")
	}

	return client.New(u, tok), nil
}

func isAPINotFound(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

func isAPIForbidden(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return false
}

func isAPIConflict(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 409
	}
	return false
}

func exitCode(err error) int {
	if isAPINotFound(err) {
		return ExitNotFound
	}
	if isAPIForbidden(err) {
		return ExitForbidden
	}
	if isAPIConflict(err) {
		return ExitConflict
	}
	return ExitError
}

func handleError(err error, errCode string, input interface{}, suggestion string, _ ...bool) {
	retryable := false
	if apiErr, ok := err.(*client.APIError); ok {
		retryable = apiErr.StatusCode >= 500
	}
	output.WriteError(errCode, err.Error(), input, suggestion, retryable)
}

// NewRootCmd builds and returns the root command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "linkding",
		Short: "CLI for the linkding bookmark manager",
		Long: `linkding is a CLI for interacting with any self-hosted linkding instance.

Configuration (flags or environment variables):
  --url / LINKDING_URL    Base URL of your linkding instance (e.g. https://links.example.com)
  --token / LINKDING_TOKEN  API token from Settings > Integrations

All commands support --json for machine-readable output and -q/--quiet for
pipe-friendly bare output. Use --help on any command to see examples.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagURL, "url", "", "Linkding base URL (or set LINKDING_URL)")
	root.PersistentFlags().StringVar(&flagToken, "token", "", "Linkding API token (or set LINKDING_TOKEN)")
	root.PersistentFlags().BoolVar(&output.IsJSONMode, "json", false, "Output machine-readable JSON to stdout")
	root.PersistentFlags().BoolVarP(&output.IsQuietMode, "quiet", "q", false, "Quiet / pipe-friendly output (one item per line, no decoration)")

	root.AddCommand(
		newBookmarkCmd(),
		newTagCmd(),
		newVersionCmd(),
	)

	return root
}
