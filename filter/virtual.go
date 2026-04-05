// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Virtual tag registry. New tags are registered via init() in virtual_*.go
// category files. This file never needs to change when adding new tags.
package filter

import (
	"sort"
	"time"

	"github.com/hugginsio/taskkit"
)

type virtualTag struct {
	positive func(now time.Time) Clause                   // SQL clause for HasTag
	negative func(now time.Time) Clause                   // SQL clause for LacksTag
	applies  func(task *taskkit.Task, now time.Time) bool // Go predicate for ComputeVirtualTags
}

var virtualTags = map[string]virtualTag{}

// register is called from init() in each virtual_*.go category file.
func register(name string, vt virtualTag) {
	virtualTags[name] = vt
}

// IsVirtualTag reports whether name is a registered virtual tag.
// The lookup is exact - virtual tag names are uppercase by convention.
func IsVirtualTag(name string) bool {
	_, ok := virtualTags[name]
	return ok
}

// ComputeVirtualTags returns the names of all virtual tags that apply to task,
// sorted alphabetically. Called at the client layer after full hydration.
func ComputeVirtualTags(task *taskkit.Task, now time.Time) []string {
	var names []string
	for name, vt := range virtualTags {
		if vt.applies(task, now) {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names
}
