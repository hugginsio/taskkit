// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package client

import (
	"errors"
	"strings"

	"github.com/hugginsio/taskkit"
)

func validate(task *taskkit.Task) error {
	if strings.TrimSpace(task.Description) == "" {
		return errors.New("description cannot be empty")
	}

	// TODO: check that the dates are in order:
	// wait must be less than or equal to scheduled,
	// scheduled must be less than or equal to deadline.

	return nil
}
