// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package format

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/hugginsio/taskkit"
)

const prettyDateLayout = "2006-01-02 15:04"

var (
	prettyHeaderStyle = lipgloss.NewStyle().Bold(true)
	prettyFaintStyle  = lipgloss.NewStyle().Faint(true)
)

type prettyFormatter struct{ columns []string }

// taskColumn defines a single column in the task table.
type taskColumn struct {
	name   string // --columns key, e.g. "blocked_by"
	header string // printed header
	show   bool
	check  func(*taskkit.Task) bool   // returns true if this task has data for the column
	value  func(*taskkit.Task) string // renders the cell value
}

// taskColumns returns the full column list in default display order.
func taskColumns() []taskColumn {
	return []taskColumn{
		{
			name: "id", header: "ID", show: true,
			value: func(t *taskkit.Task) string { return fmt.Sprintf("%d", t.DisplayID) },
		},
		{
			name: "blocked_by", header: "Blocked By",
			check: func(t *taskkit.Task) bool { return len(t.BlockedBy) > 0 },
			value: func(t *taskkit.Task) string { return taskIDList(t.BlockedBy) },
		},
		{
			name: "blocking", header: "Blocking",
			check: func(t *taskkit.Task) bool { return len(t.Blocking) > 0 },
			value: func(t *taskkit.Task) string { return taskIDList(t.Blocking) },
		},
		{
			name: "project", header: "Project",
			check: func(t *taskkit.Task) bool { return t.Project != "" },
			value: func(t *taskkit.Task) string { return t.Project },
		},
		{
			name: "tags", header: "Tags",
			check: func(t *taskkit.Task) bool {
				return len(t.Tags) > 0
			},
			value: func(t *taskkit.Task) string {
				return strings.Join(t.Tags, ", ")
			},
		},
		{
			name: "status", header: "Status",
			check: func(t *taskkit.Task) bool { return t.Status != taskkit.StatusPending },
			value: func(t *taskkit.Task) string { return string(t.Status) },
		},
		{
			name: "wait", header: "Wait",
			check: func(t *taskkit.Task) bool { return t.Wait != nil },
			value: func(t *taskkit.Task) string { return prettyDate(t.Wait) },
		},
		{
			name: "scheduled", header: "Scheduled",
			check: func(t *taskkit.Task) bool { return t.Scheduled != nil },
			value: func(t *taskkit.Task) string { return prettyDate(t.Scheduled) },
		},
		{
			name: "deadline", header: "Deadline",
			check: func(t *taskkit.Task) bool { return t.Deadline != nil },
			value: func(t *taskkit.Task) string { return prettyDate(t.Deadline) },
		},
		{
			name: "description", header: "Description", show: true,
			value: func(t *taskkit.Task) string { return t.Description },
		},
		{
			name: "urgency", header: "Urgency", show: true,
			value: func(t *taskkit.Task) string { return fmt.Sprintf("%.2f", t.Urgency) },
		},
	}
}

func (f *prettyFormatter) Tasks(w io.Writer, tasks []*taskkit.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks.")
		return nil
	}

	cols := taskColumns()

	if len(f.columns) > 0 {
		cols = selectColumns(cols, f.columns)
	} else {
		for _, t := range tasks {
			for i := range cols {
				if !cols[i].show && cols[i].check != nil && cols[i].check(t) {
					cols[i].show = true
				}
			}
		}
	}

	var headers []string
	for _, c := range cols {
		if c.show {
			headers = append(headers, c.header)
		}
	}

	rows := make([][]string, len(tasks))
	for i, t := range tasks {
		for _, c := range cols {
			if c.show {
				rows[i] = append(rows[i], c.value(t))
			}
		}
	}

	tbl := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return prettyHeaderStyle
			}
			return lipgloss.NewStyle()
		}).
		Headers(headers...).
		Rows(rows...)

	fmt.Fprintln(w, tbl.Render())
	return nil
}

func (f *prettyFormatter) Task(w io.Writer, task *taskkit.Task) error {
	type row struct{ label, value string }
	candidates := []row{
		{"ID", strconv.Itoa(task.DisplayID)},
		{"Description", task.Description},
		{"Status", string(task.Status)},
		{"Project", task.Project},
		{"Blocking", taskIDList(task.Blocking)},
		{"Blocked By", taskIDList(task.BlockedBy)},
		{"Created", task.Created.Local().Format(prettyDateLayout)},
		{"Wait", prettyDate(task.Wait)},
		{"Scheduled", prettyDate(task.Scheduled)},
		{"Deadline", prettyDate(task.Deadline)},
		{"Modified", task.Modified.Local().Format(prettyDateLayout)},
		{"Tags", strings.Join(task.Tags, ", ")},
		{"Virtual Tags", strings.Join(task.VirtualTags, ", ")},
		{"ULID", task.TaskID.String()},
	}

	candidates = append(candidates, row{"Urgency", fmt.Sprintf("%.2f", task.Urgency)})

	var rows [][]string
	for _, r := range candidates {
		if r.value != "" {
			rows = append(rows, []string{r.label, r.value})
		}
	}

	tbl := table.New().
		StyleFunc(func(r, col int) lipgloss.Style {
			if col == 0 {
				return prettyFaintStyle
			}
			return lipgloss.NewStyle()
		}).
		Rows(rows...)

	fmt.Fprintln(w, tbl.Render())

	if len(task.History) > 0 {
		renderHistory(w, task.History)
	}

	return nil
}

func (f *prettyFormatter) ProjectSummaries(w io.Writer, summaries []*taskkit.ProjectSummary) error {
	if len(summaries) == 0 {
		fmt.Fprintln(w, "No projects.")
		return nil
	}

	rows := make([][]string, len(summaries))
	for i, s := range summaries {
		total := s.Remaining + s.Completed
		rows[i] = []string{
			s.Name,
			fmt.Sprintf("%d", total),
			fmt.Sprintf("%d", s.Remaining),
			fmt.Sprintf("%d", s.Completed),
			fmt.Sprintf("%d%%", s.Percent),
		}
	}

	tbl := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return prettyHeaderStyle
			}

			return lipgloss.NewStyle()
		}).
		Headers("Project", "Total", "Remaining", "Completed", "%").
		Rows(rows...)

	fmt.Fprintln(w, tbl.Render())
	return nil
}

// selectColumns returns columns filtered and ordered to match the requested names.
// Unknown names are silently ignored.
func selectColumns(all []taskColumn, names []string) []taskColumn {
	byName := make(map[string]taskColumn, len(all))
	for _, c := range all {
		byName[c.name] = c
	}

	var out []taskColumn
	for _, name := range names {
		if c, ok := byName[name]; ok {
			c.show = true
			out = append(out, c)
		}
	}

	return out
}

// taskIDList returns a comma-separated string of display IDs for a slice of tasks.
func taskIDList(tasks []*taskkit.Task) string {
	parts := make([]string, len(tasks))
	for i, t := range tasks {
		parts[i] = fmt.Sprintf("%d", t.DisplayID)
	}

	return strings.Join(parts, ", ")
}

func prettyDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Local().Format(prettyDateLayout)
}

func (f *prettyFormatter) UrgencyBreakdown(w io.Writer, task *taskkit.Task) error {
	if len(task.UrgencyBreakdown) == 0 {
		fmt.Fprintln(w, "No urgency components.")
		return nil
	}

	fmtNum := func(v float64) string { return fmt.Sprintf("%.3g", v) }
	empty := []string{"", "", "", "", "", ""}

	rows := make([][]string, 0, len(task.UrgencyBreakdown)+2)
	for _, c := range task.UrgencyBreakdown {
		rows = append(rows, []string{
			c.Label,
			fmtNum(c.Coefficient),
			"*",
			fmtNum(c.Weight),
			"=",
			fmtNum(c.Coefficient * c.Weight),
		})
	}

	sepRow := append([]string{}, empty...)
	sepRow[5] = strings.Repeat("-", 5)
	rows = append(rows, sepRow)

	totalRow := append([]string{}, empty...)
	totalRow[5] = strconv.FormatFloat(task.Urgency, 'f', -1, 64)
	rows = append(rows, totalRow)

	sepIdx := len(rows) - 2
	totalIdx := len(rows) - 1

	tbl := table.New().
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderHeader(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			pad := col < 5
			switch {
			case row == sepIdx || row == totalIdx:
				s := lipgloss.NewStyle().Bold(row == totalIdx)
				if pad {
					s = s.PaddingRight(1)
				}

				return s
			case col == 0:
				return prettyFaintStyle.PaddingRight(2)
			case col == 1 || col == 3 || col == 5:
				s := lipgloss.NewStyle().AlignHorizontal(lipgloss.Right)
				if pad {
					s = s.PaddingRight(1)
				}

				return s
			case col == 2 || col == 4:
				s := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
				if pad {
					s = s.PaddingRight(1)
				}

				return s
			}

			return lipgloss.NewStyle()
		}).
		Rows(rows...)

	fmt.Fprintln(w, tbl.Render())
	return nil
}

const historyDateLayout = "2006-01-02 15:04:05"

// renderHistory prints an audit log using a two-column table. Each
// operation's timestamp appears in the first row; subsequent changes
// in the same operation leave the date column blank.
func renderHistory(w io.Writer, entries []taskkit.HistoryEntry) {
	var rows [][]string
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		lines := historyLines(entry)
		ts := entry.Time.Local().Format(historyDateLayout)
		for i, line := range lines {
			date := ts
			if i > 0 {
				date = ""
			}

			rows = append(rows, []string{date, line})
		}
	}

	tbl := table.New().
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return prettyHeaderStyle
			}

			if col == 0 {
				return prettyFaintStyle
			}

			return lipgloss.NewStyle()
		}).
		Headers("Date", "Modification").
		Rows(rows...)

	fmt.Fprintln(w, tbl.Render())
}

// historyLines returns one human-readable line per change in the entry.
// Returns nil for create entries, which are skipped by the caller.
func historyLines(entry taskkit.HistoryEntry) []string {
	if entry.Action == "create" {
		return nil
	}

	lines := make([]string, 0, len(entry.Changes))
	for _, c := range entry.Changes {
		if s := formatChange(c); s != "" {
			lines = append(lines, s)
		}
	}

	if len(lines) == 0 {
		return []string{entry.Action + "."}
	}

	return lines
}

// formatChange converts a HistoryChange into a readable sentence.
func formatChange(c taskkit.HistoryChange) string {
	switch c.Field {
	case "description":
		return fmt.Sprintf("Description set to '%s'.", c.NewValue)
	case "status":
		return fmt.Sprintf("Status changed from '%s' to '%s'.", c.OldValue, c.NewValue)
	case "project":
		if c.NewValue == "" {
			return "Project cleared."
		}
		return fmt.Sprintf("Project set to '%s'.", c.NewValue)
	case "deadline":
		if c.NewValue == "" {
			return "Deadline cleared."
		}
		return fmt.Sprintf("Deadline set to '%s'.", prettyRFC3339(c.NewValue))
	case "scheduled":
		if c.NewValue == "" {
			return "Scheduled date cleared."
		}
		return fmt.Sprintf("Scheduled set to '%s'.", prettyRFC3339(c.NewValue))
	case "wait":
		if c.NewValue == "" {
			return "Wait date cleared."
		}
		return fmt.Sprintf("Wait set to '%s'.", prettyRFC3339(c.NewValue))
	case "tag.add":
		return fmt.Sprintf("Tag '%s' added.", c.NewValue)
	case "tag.remove":
		return fmt.Sprintf("Tag '%s' removed.", c.OldValue)
	case "dependency.add":
		return fmt.Sprintf("Blocked-by dependency added (%s).", c.NewValue)
	case "dependency.remove":
		return fmt.Sprintf("Blocked-by dependency removed (%s).", c.OldValue)
	case "display_id":
		return "" // internal; skip
	default:
		return fmt.Sprintf("%s changed.", c.Field)
	}
}

// prettyRFC3339 parses an RFC3339 timestamp string and formats it for display.
func prettyRFC3339(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Local().Format(prettyDateLayout)
}
