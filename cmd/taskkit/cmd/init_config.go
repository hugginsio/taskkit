// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hugginsio/taskkit/config"
)

// promptCreateConfig asks the user whether to create a default config file at
// path. Returns true if the file was successfully created.
func promptCreateConfig(in io.Reader, out io.Writer, path string) bool {
	fmt.Fprintf(out, "No configuration file found.\nCreate default config at %s? [Y/n] ", path)

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil {
		return false
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	if answer != "" && answer != "y" && answer != "yes" {
		return false
	}

	created, err := config.CreateDefault()
	if err != nil {
		fmt.Fprintf(out, "Error: %v\n", err)
		return false
	}

	fmt.Fprintf(out, "Created %s\n", created)
	return true
}
