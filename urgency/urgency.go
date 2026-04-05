// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Package urgency computes a weighted urgency score for tasks.
// Scores are a polynomial sum of configurable coefficients applied to task
// attributes. Higher scores indicate more urgent tasks.
package urgency

import (
	"math"
	"time"

	"github.com/hugginsio/taskkit"
)

// Weights holds the urgency scoring coefficients.
// Defaults for all fields are defined in config/default.yaml.
type Weights struct {
	Deadline  float64 `yaml:"deadline"`
	Scheduled float64 `yaml:"scheduled"`
	Age       float64 `yaml:"age"`
	AgeNorm   float64 `yaml:"age_norm"`
	Tags      float64 `yaml:"tags"`
	Waiting   float64 `yaml:"waiting"`
	Blocked   float64 `yaml:"blocked"`
	Blocking  float64 `yaml:"blocking"`
}

// Score computes the urgency score for task at the given reference time using w.
// now is injected so callers and tests can use a fixed clock.
func Score(task *taskkit.Task, w Weights, now time.Time) float64 {
	var score float64

	// Deadline.
	if task.Deadline != nil {
		score += w.Deadline * dueFactor(now, *task.Deadline)
	}

	// Scheduled — past start date means work should have begun.
	if task.Scheduled != nil && task.Scheduled.Before(now) {
		score += w.Scheduled
	}

	// Age — derived from the creation timestamp already on the domain type.
	score += w.Age * ageFactor(now, task.Created, w.AgeNorm)

	// Tags.
	if n := len(task.Tags); n > 0 {
		score += w.Tags * tagModifier(n)
	}

	// Waiting penalty.
	if task.Status == taskkit.StatusWaiting {
		score += w.Waiting
	}

	// Blocked — has active blocked-by dependencies.
	if len(task.BlockedBy) > 0 {
		score += w.Blocked
	}

	// Blocking — this task is blocking other tasks.
	if len(task.Blocking) > 0 {
		score += w.Blocking
	}

	return math.Round(score*100) / 100
}

// Components returns the non-zero urgency terms for a task, suitable for
// rendering a breakdown table. Only terms whose weight * coefficient != 0
// are included.
func Components(task *taskkit.Task, w Weights, now time.Time) []taskkit.UrgencyComponent {
	type term struct {
		label string
		coeff float64
		w     float64
	}

	var terms []term

	if task.Deadline != nil {
		terms = append(terms, term{"due", dueFactor(now, *task.Deadline), w.Deadline})
	}

	if task.Scheduled != nil && task.Scheduled.Before(now) {
		terms = append(terms, term{"scheduled", 1, w.Scheduled})
	}

	terms = append(terms, term{"age", ageFactor(now, task.Created, w.AgeNorm), w.Age})

	if n := len(task.Tags); n > 0 {
		terms = append(terms, term{"tags", tagModifier(n), w.Tags})
	}

	if task.Status == taskkit.StatusWaiting {
		terms = append(terms, term{"waiting", 1, w.Waiting})
	}

	if len(task.BlockedBy) > 0 {
		terms = append(terms, term{"blocked", 1, w.Blocked})
	}

	if len(task.Blocking) > 0 {
		terms = append(terms, term{"blocking", 1, w.Blocking})
	}

	var out []taskkit.UrgencyComponent
	for _, t := range terms {
		if t.coeff*t.w != 0 {
			out = append(out, taskkit.UrgencyComponent{
				Label:       t.label,
				Coefficient: t.coeff,
				Weight:      t.w,
			})
		}
	}

	return out
}

// ScoreAll populates the Urgency field on each task in place and returns the
// same slice. Callers are responsible for sorting if desired.
func ScoreAll(tasks []*taskkit.Task, w Weights, now time.Time) []*taskkit.Task {
	for _, task := range tasks {
		task.Urgency = Score(task, w, now)
	}

	return tasks
}

// dueFactor returns a value in [0.0, 1.0] representing how urgently the
// deadline contributes to the score. Overdue tasks return 1.0; tasks due
// further than 14 days away return 0.0.
func dueFactor(now, deadline time.Time) float64 {
	days := deadline.Sub(now).Hours() / 24.0
	switch {
	case days < 0:
		return 1.0
	case days < 7:
		return 0.1 + (7.0-days)/7.0*0.9
	case days < 14:
		return (14.0 - days) / 7.0 * 0.2
	default:
		return 0.0
	}
}

// ageFactor returns a value in [0.0, 1.0] representing how the task's age
// contributes to urgency. Saturates at 1.0 after normDays days.
func ageFactor(now, created time.Time, normDays float64) float64 {
	age := now.Sub(created).Hours() / 24.0
	return math.Min(age/normDays, 1.0)
}

// tagModifier returns a scaling factor based on the number of tags:
// 0.8 for one tag, 0.9 for two, 1.0 for three or more.
func tagModifier(n int) float64 {
	switch n {
	case 1:
		return 0.8
	case 2:
		return 0.9
	default:
		return 1.0
	}
}
