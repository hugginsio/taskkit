// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

var ulidToDateCmd = &cobra.Command{
	Use:     "date-from-ulid <ulid>",
	Short:   "Extract a date from a ULID",
	Args:    cobra.ExactArgs(1),
	GroupID: "utility",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := ulid.ParseStrict(strings.TrimSpace(args[0]))
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), id.Timestamp().Format(time.RFC3339))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ulidToDateCmd)
}
