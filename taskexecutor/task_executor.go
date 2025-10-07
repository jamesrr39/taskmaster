package taskexecutor

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jamesrr39/taskmaster/taskrunner"
)

const TASKRUNNER_SOURCE_NAME string = "TASKRUNNER"

func ExecuteJobRun(taskRun *taskrunner.TaskRun, taskRunStatusChangeChan chan *taskrunner.TaskRun, logFile io.Writer, workspaceDir string, providesNow NowProvider) error {
	panic("write to temp dir")
	scriptFilePath := filepath.Join(workspaceDir, "script")
	err := os.WriteFile(scriptFilePath, []byte(taskRun.Task.Script), 0500)
	if nil != err {
		return handleTaskrunnerError("Couldn't prepare and move to workspace. Error: "+err.Error(), logFile, taskRunStatusChangeChan, taskRun, providesNow)
	}

	cmd := exec.Command(scriptFilePath)
	stdoutPipe, err := cmd.StdoutPipe()
	if nil != err {
		return handleTaskrunnerError("Couldn't obtain stdoutpipe. Error: "+err.Error(), logFile, taskRunStatusChangeChan, taskRun, providesNow)
	}

	stderrPipe, err := cmd.StderrPipe()
	if nil != err {
		return handleTaskrunnerError("Couldn't obtain stderrpipe. Error: "+err.Error(), logFile, taskRunStatusChangeChan, taskRun, providesNow)
	}

	go writeToLogFile(stdoutPipe, logFile, "STDOUT", providesNow)
	go writeToLogFile(stderrPipe, logFile, "STDERR", providesNow)

	err = cmd.Start()
	if nil != err {
		return handleTaskrunnerError("Couldn't start script. Error: "+err.Error(), logFile, taskRunStatusChangeChan, taskRun, providesNow)

	}

	err = cmd.Wait()
	if nil != err {
		switch err.(type) {
		case *exec.ExitError:
			taskRun.State = taskrunner.JOB_RUN_STATE_FAILED
		default:
			taskRun.State = taskrunner.JOB_RUN_STATE_UNKNOWN
		}
	} else {
		taskRun.State = taskrunner.JOB_RUN_STATE_SUCCESS
	}
	taskRun.EndTimestamp = time.Now().Unix()
	taskRunStatusChangeChan <- taskRun

	return nil
}

func handleTaskrunnerError(errorMessage string, logFile io.Writer, jobRunStateChan chan *taskrunner.TaskRun, jobRun *taskrunner.TaskRun, providesNow NowProvider) error {
	jobRun.EndTimestamp = providesNow().Unix()
	jobRun.State = taskrunner.JOB_RUN_STATE_FAILED
	jobRunStateChan <- jobRun
	err := writeStringToLogFile(errorMessage, logFile, TASKRUNNER_SOURCE_NAME, providesNow)
	if nil != err {
		return err
	}
	return nil
}
