package taskexecutor

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jamesrr39/taskmaster/taskrunner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExecuteJobRun(t *testing.T) {
	workspaceDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(workspaceDir)
		if nil != err {
			t.Errorf("Couldn't remove the tempdir at '%s'. Error: %s.\n", workspaceDir, err)
		} else {
			t.Logf("Successfully removed workspace dir at %s\n", workspaceDir)
		}
	}()

	logFile := bytes.NewBuffer(nil)

	jobRunStateChan := make(chan *taskrunner.TaskRun)

	task, err := taskrunner.NewTask("my task", "the big task", "#!/bin/bash\n\necho 'task failed'\nexit 1")
	require.NoError(t, err)
	taskRun := task.NewTaskRun(1, taskrunner.Timestamp(mockNowProvider()))

	jobRunStateGoRoutineDoneChan := make(chan bool)

	var newJobRunState *taskrunner.TaskRun
	go func() {
		newJobRunState = <-jobRunStateChan
		jobRunStateGoRoutineDoneChan <- true
	}()

	err = ExecuteJobRun(task, taskRun, jobRunStateChan, logFile, workspaceDir, mockNowProvider)
	require.NoError(t, err)

	<-jobRunStateGoRoutineDoneChan
	assert.Equal(t, "03:04:05.006: STDOUT: task failed\n", string(logFile.Bytes()))
	assert.Equal(t, taskrunner.JOB_RUN_STATE_FAILED, newJobRunState.State)

}

func Test_handleTaskrunnerError(t *testing.T) {
	errorMessage := "setup failed"
	logFile := bytes.NewBuffer(nil)
	jobRunStateChan := make(chan *taskrunner.TaskRun)

	task, err := taskrunner.NewTask("my task", "the big task", "#!/bin/bash\n\necho 'my task'")
	require.NoError(t, err)
	jobRun := task.NewTaskRun(1, taskrunner.Timestamp(mockNowProvider()))

	var newJobRunState *taskrunner.TaskRun
	go func() {
		// listen to jobRunStateChan
		newJobRunState = <-jobRunStateChan
	}()

	err = handleTaskrunnerError(errorMessage, logFile, jobRunStateChan, jobRun, mockNowProvider)
	require.NoError(t, err)

	assert.Equal(t, taskrunner.JOB_RUN_STATE_FAILED, newJobRunState.State)
	assert.Equal(t, int64(946782245), newJobRunState.EndTimestamp)

	assert.Equal(t, "03:04:05.006: TASKRUNNER: setup failed\n", string(logFile.Bytes()))
}
