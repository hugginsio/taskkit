// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

// DescriptionIs matches tasks whose description exactly equals s.
func DescriptionIs(s string) Filter {
	return func() Clause {
		return Clause{SQL: "description = ?", Params: []any{s}}
	}
}

// DescriptionContains matches tasks whose description contains s (case-insensitive).
func DescriptionContains(s string) Filter {
	return func() Clause {
		return Clause{SQL: "description LIKE ?", Params: []any{"%" + s + "%"}}
	}
}

// Project matches tasks belonging to the named project.
func Project(name string) Filter {
	return func() Clause {
		return Clause{SQL: "project = ?", Params: []any{name}}
	}
}

// ProjectPrefix matches tasks whose project name starts with prefix.
func ProjectPrefix(prefix string) Filter {
	return func() Clause {
		return Clause{SQL: "project LIKE ?", Params: []any{prefix + "%"}}
	}
}

// HasProject matches tasks that have any project set.
func HasProject() Filter {
	return func() Clause {
		return Clause{SQL: "project IS NOT NULL"}
	}
}

// NoProject matches tasks with no project set.
func NoProject() Filter {
	return func() Clause {
		return Clause{SQL: "project IS NULL"}
	}
}
