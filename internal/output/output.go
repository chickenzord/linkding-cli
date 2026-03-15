package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// IsJSONMode returns true if --json was requested.
var IsJSONMode bool

// IsQuietMode returns true if --quiet was requested.
var IsQuietMode bool

// JSON writes v as JSON to stdout.
func JSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		Errf("failed to encode JSON: %v\n", err)
	}
}

// Infof writes a message to stderr in JSON mode, stdout in human mode.
func Infof(format string, args ...interface{}) {
	if IsJSONMode {
		fmt.Fprintf(os.Stderr, format, args...)
	} else {
		fmt.Fprintf(os.Stdout, format, args...)
	}
}

// Errf writes an error message to stderr.
func Errf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// ErrorJSON writes a structured error to stdout and a human message to stderr.
type ErrorPayload struct {
	Error      string      `json:"error"`
	Message    string      `json:"message"`
	Input      interface{} `json:"input,omitempty"`
	Suggestion string      `json:"suggestion,omitempty"`
	Retryable  bool        `json:"retryable"`
}

func WriteError(code string, message string, input interface{}, suggestion string, retryable bool) {
	if IsJSONMode {
		payload := ErrorPayload{
			Error:      code,
			Message:    message,
			Input:      input,
			Suggestion: suggestion,
			Retryable:  retryable,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
	}
	msg := "Error: " + message
	if suggestion != "" {
		msg += "\nHint: " + suggestion
	}
	fmt.Fprintln(os.Stderr, msg)
}

// NewTabWriter returns a tabwriter writing to stdout.
func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}
