// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var viewConfigCmd = &cobra.Command{
	Use:     "view-config",
	Short:   "Print the resolved configuration",
	GroupID: CommandGroupUtility,
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), string(out))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewConfigCmd)
}
