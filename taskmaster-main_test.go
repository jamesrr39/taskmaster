package main

import (
	"testing"
	"time"

	"github.com/jamesrr39/taskmaster/taskrunner"
	"github.com/stretchr/testify/assert"
)

func Test_sortEntries(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		entries []ListTaskEntry
		want    []ListTaskEntry
	}{
		{
			name: "2 nil latest run entries",
			entries: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #1"},
				}, {
					Task: &taskrunner.Task{Name: "task #2"},
				}, {
					Task: &taskrunner.Task{Name: "task #3"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				},
			},
			want: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #3"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "task #1"},
				}, {
					Task: &taskrunner.Task{Name: "task #2"},
				},
			},
		}, {
			name: "1 nil latest run entries, non-nil should be shown first",
			entries: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #1"},
				}, {
					Task: &taskrunner.Task{Name: "task #2"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				},
			},
			want: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #2"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "task #1"},
				},
			},
		}, {
			name: "2 tasks with run entries",
			entries: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #1"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "task #3"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				},
			},
			want: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #3"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(1, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "task #1"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				},
			},
		}, {}, {
			name: "2 tasks with run entries starting at the same time, should be sorted by name",
			entries: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "task #2"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "Task #1"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				},
			},
			want: []ListTaskEntry{
				{
					Task: &taskrunner.Task{Name: "Task #1"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				}, {
					Task: &taskrunner.Task{Name: "task #2"},
					LatestRuns: []*taskrunner.TaskRun{
						{StartTimestamp: taskrunner.Timestamp(time.Unix(0, 0))},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sortEntries(tt.entries), tt.name)
		})
	}
}
