// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	goversion "github.com/caarlos0/go-version"
	"github.com/charmbracelet/fang"
	"github.com/hugginsio/taskkit/client"
	"github.com/hugginsio/taskkit/config"
	"github.com/spf13/cobra"
)

var c *client.Client   // The TaskKit Client, instantiated at runtime.
var cfg *config.Config // The resolved configuration, set alongside c.

var rootCmd = &cobra.Command{
	Use:   "taskkit",
	Short: "A toolkit for task and project management",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
		}
	},
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
	// TODO: more groups
	rootCmd.AddGroup(&cobra.Group{ID: "utility", Title: "Utilities"})
}
