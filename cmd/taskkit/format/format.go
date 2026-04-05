// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Package format provides output formatters for task data.
package format

import (
	"fmt"
	"io"

	"github.com/hugginsio/taskkit"
	"github.com/spf13/cobra"
)

// Formatter writes task output to a writer.
type Formatter interface {
	Task(w io.Writer, task *taskkit.Task) error
	Tasks(w io.Writer, tasks []*taskkit.Task) error
	ProjectSummaries(w io.Writer, summaries []*taskkit.ProjectSummary) error
	UrgencyBreakdown(w io.Writer, task *taskkit.Task) error
}

// NewFormatter returns the Formatter for the given format string and optional
// column list (used by the pretty formatter; ignored by json).
func NewFormatter(f string, columns []string) (Formatter, error) {
	switch f {
	case "json":
		return &jsonFormatter{}, nil
	case "pretty":
		return &prettyFormatter{columns: columns}, nil
	default:
		return nil, fmt.Errorf("unknown output format %q", f)
	}
}

// NewFormatterFromCmd returns the Formatter based on the --output flag value.
// If --columns is registered on the command, its value is forwarded to the formatter.
func NewFormatterFromCmd(cmd *cobra.Command) (Formatter, error) {
	f, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("output format: %w", err)
	}

	var cols []string
	if cmd.Flags().Lookup("columns") != nil {
		cols, _ = cmd.Flags().GetStringSlice("columns")
	}

	return NewFormatter(f, cols)
}
