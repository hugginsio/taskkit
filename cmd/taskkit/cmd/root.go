// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	goversion "github.com/caarlos0/go-version"
	"github.com/charmbracelet/fang"
	"github.com/hugginsio/taskkit/client"
	"github.com/hugginsio/taskkit/config"
	"github.com/spf13/cobra"
)

var c *client.Client        // The TaskKit Client, instantiated at runtime.
var cfg *config.Config      // The resolved configuration, set alongside c.
var handleError func(error) // Fang-styled error printer, set in Execute.

var CommandGroupTask = "task"
var CommandGroupUtility = "utility"
var CommandGroupReport = "report"

var rootCmd = &cobra.Command{
	Use:   "taskkit",
	Short: "A toolkit for task and project management",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Cobra's internal completion commands initialize the client themselves
		// only if their completion func needs it (via ensureClient).
		if n := cmd.Name(); n == "__complete" || n == "__completeNoDesc" {
			return nil
		}

		// cfg is pre-loaded in Execute; reload only if something went wrong.
		if cfg == nil {
			loaded, err := config.Load()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}

			cfg = loaded
		}

		taskkitClient, err := client.NewClient(cmd.Context(), cfg)
		if err != nil {
			return fmt.Errorf("client: %w", err)
		}

		c = taskkitClient
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if c != nil {
			c.Close()
			c = nil
		}
	},
}

// initErrorHandler constructs a fang-styled error printer for use in the REPL.
// It replicates the subset of fang's internal style setup needed for errors.
func initErrorHandler() {
	cs := fang.DefaultColorScheme(lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)))
	styles := fang.Styles{
		ErrorHeader: lipgloss.NewStyle().
			Foreground(cs.ErrorHeader[0]).
			Background(cs.ErrorHeader[1]).
			Bold(true).
			Padding(0, 1).
			Margin(1).
			MarginLeft(2).
			SetString("ERROR"),
		ErrorText: lipgloss.NewStyle().MarginLeft(2),
	}

	handleError = func(err error) {
		fang.DefaultErrorHandler(rootCmd.ErrOrStderr(), styles, err)
	}
}

func Execute() {
	// Load config before cobra parses so we can register report shortcuts based
	// on the resolved config (including any user overrides or removals).
	loaded, err := config.Load()
	if errors.Is(err, config.ErrNotFound) {
		if created := promptCreateConfig(os.Stdin, os.Stderr, config.DefaultPath()); created {
			loaded, _ = config.Load()
		}
	}

	if loaded != nil {
		cfg = loaded
		registerReportShortcuts(cfg)
	}

	initErrorHandler()

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithCommit(goversion.GetVersionInfo().GitCommit),
		fang.WithVersion(goversion.GetVersionInfo().GitVersion),
	); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: CommandGroupTask, Title: "Tasks"})
	rootCmd.AddGroup(&cobra.Group{ID: CommandGroupReport, Title: "Reports"})
	rootCmd.AddGroup(&cobra.Group{ID: CommandGroupUtility, Title: "Utilities"})

	rootCmd.SetCompletionCommandGroupID(CommandGroupUtility)
	rootCmd.SetHelpCommandGroupID(CommandGroupUtility)
}
