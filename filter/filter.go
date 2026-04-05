// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Package filter provides composable SQL clause builders for querying tasks.
// Filters compose with AND semantics by default; use Or for OR semantics.
package filter

import "strings"

// Clause holds a SQL fragment and its bound parameters.
type Clause struct {
	SQL    string
	Params []any
}

// Filter is a function that returns a Clause.
type Filter func() Clause

// Or combines one or more filters with OR semantics.
func Or(filters ...Filter) Filter {
	return func() Clause {
		var parts []string
		var params []any
		for _, f := range filters {
			c := f()
			if c.SQL != "" {
				parts = append(parts, c.SQL)
				params = append(params, c.Params...)
			}
		}

		if len(parts) == 0 {
			return Clause{}
		}

		return Clause{
			SQL:    "(" + strings.Join(parts, " OR ") + ")",
			Params: params,
		}
	}
}

// BuildTaskQuery assembles a full SELECT against the tasks table with all
// provided filters ANDed together. With no filters it returns all tasks.
func BuildTaskQuery(filters ...Filter) (string, []any) {
	const base = "SELECT task_id, display_id, project, description, status, deadline, scheduled, wait FROM tasks"

	var clauses []string
	var params []any

	for _, f := range filters {
		c := f()
		if c.SQL != "" {
			clauses = append(clauses, c.SQL)
			params = append(params, c.Params...)
		}
	}

	if len(clauses) == 0 {
		return base, nil
	}

	return base + " WHERE " + strings.Join(clauses, " AND "), params
}

// inClause builds a "field IN (?, ?, ...)" fragment.
func inClause(field string, values []any) Clause {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	return Clause{
		SQL:    field + " IN (" + placeholders + ")",
		Params: values,
	}
}
