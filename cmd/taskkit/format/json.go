// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package format

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hugginsio/taskkit"
)

type jsonFormatter struct{}

func (f *jsonFormatter) Task(w io.Writer, task *taskkit.Task) error {
	return f.encode(w, toLocalTask(task))
}

func (f *jsonFormatter) Tasks(w io.Writer, tasks []*taskkit.Task) error {
	local := make([]*taskkit.Task, len(tasks))
	for i, t := range tasks {
		local[i] = toLocalTask(t)
	}

	return f.encode(w, local)
}

func (f *jsonFormatter) ProjectSummaries(w io.Writer, summaries []*taskkit.ProjectSummary) error {
	return f.encode(w, summaries)
}

func (f *jsonFormatter) UrgencyBreakdown(w io.Writer, task *taskkit.Task) error {
	return f.encode(w, task.UrgencyBreakdown)
}

// toLocalTask returns a shallow copy of task with all time.Time fields
// converted to the local timezone so JSON output reflects local time.
func toLocalTask(t *taskkit.Task) *taskkit.Task {
	copy := *t
	copy.Created = copy.Created.In(time.Local)
	copy.Modified = copy.Modified.In(time.Local)
	if copy.Deadline != nil {
		local := copy.Deadline.In(time.Local)
		copy.Deadline = &local
	}

	if copy.Scheduled != nil {
		local := copy.Scheduled.In(time.Local)
		copy.Scheduled = &local
	}

	if copy.Wait != nil {
		local := copy.Wait.In(time.Local)
		copy.Wait = &local
	}

	return &copy
}

func (f *jsonFormatter) encode(w io.Writer, v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(out))
	return err
}
