// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"fmt"
	"strings"

	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/hugginsio/taskkit/config"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:     "report [name]",
	Aliases: []string{"r"},
	Short:   "List or run reports",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		flags, ok := cfg.Reports[args[0]]
		if !ok {
			return fmt.Errorf("unknown report %q; run 'taskkit report' to list available reports", args[0])
		}

		return runReportFlags(cmd, flags)
	},
}

// runReportFlags parses flagStr into a temporary command, builds filters, and
// prints the matching tasks. The output and sort flags are read from cmd.
func runReportFlags(cmd *cobra.Command, flagStr string) error {
	tmp := &cobra.Command{}
	registerFilterFlags(tmp)
	registerSortFlag(tmp)

	if err := tmp.Flags().Parse(strings.Fields(flagStr)); err != nil {
		return fmt.Errorf("report flags: %w", err)
	}

	// Allow --sort on the real command to override the report's default.
	sortBy, _ := tmp.Flags().GetString("sort")
	if cmd.Flags().Changed("sort") {
		sortBy, _ = cmd.Flags().GetString("sort")
	}

	filters, err := filtersFromFlags(cmd.Context(), tmp)
	if err != nil {
		return err
	}

	tasks, err := c.Get(cmd.Context(), filters...)
	if err != nil {
		return fmt.Errorf("report: %w", err)
	}

	sortTasks(tasks, sortBy)

	f, err := cmdformat.NewFormatterFromCmd(cmd)
	if err != nil {
		return err
	}

	return f.Tasks(cmd.OutOrStdout(), tasks)
}

// makeReportSubCmd creates a real cobra subcommand for a named built-in report.
func makeReportSubCmd(name, flagStr string) *cobra.Command {
	sub := &cobra.Command{
		Use:   name,
		Short: "Alias for " + flagStr,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReportFlags(cmd, flagStr)
		},
	}

	registerOutputFlag(sub)
	registerSortFlag(sub)

	return sub
}

// mergedReports returns the built-in defaults overlaid with user-defined reports.
// registerReportShortcuts adds a top-level command for each report in cfg.
// Called from Execute() after config is loaded so the set reflects the resolved config.
func registerReportShortcuts(cfg *config.Config) {
	for name, flags := range cfg.Reports {
		name, flags := name, flags
		shortcut := &cobra.Command{
			Use:   name,
			Short: fmt.Sprintf("Alias for 'taskkit report %s'", name),
			RunE: func(cmd *cobra.Command, _ []string) error {
				return runReportFlags(cmd, flags)
			},
		}

		registerOutputFlag(shortcut)
		registerSortFlag(shortcut)
		rootCmd.AddCommand(shortcut)
	}
}

func init() {
	registerOutputFlag(reportCmd)
	registerSortFlag(reportCmd)

	for name, flags := range config.DefaultReports {
		reportCmd.AddCommand(makeReportSubCmd(name, flags))
	}

	rootCmd.AddCommand(reportCmd)
}
