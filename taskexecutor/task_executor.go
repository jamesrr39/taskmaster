package taskexecutor

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/taskrunner"
)

func ExecuteJobRun(task *taskrunner.Task, taskRun *taskrunner.TaskRun, taskRunStatusChangeChan chan *taskrunner.TaskRun, logFile io.Writer, workspaceDir string, providesNow NowProvider) error {
	slog.Info("running job", "workspaceDir", workspaceDir, "taskRun", taskRun)

	scriptFilePath := filepath.Join(workspaceDir, "script")
	err := os.WriteFile(scriptFilePath, []byte(task.Script), 0500)
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

	go writeToLogFile(stdoutPipe, logFile, SourceTaskmasterStdout, providesNow)
	go writeToLogFile(stderrPipe, logFile, SourceTaskmasterStderr, providesNow)

	err = cmd.Start()
	if nil != err {
		return handleTaskrunnerError("Couldn't start script. Error: "+err.Error(), logFile, taskRunStatusChangeChan, taskRun, providesNow)
	}

	err = cmd.Wait()
	if nil != err {
		switch exitErr := err.(type) {
		case *exec.ExitError:
			taskRun.State = taskrunner.JOB_RUN_STATE_FAILED
			exitCode := exitErr.ExitCode()
			taskRun.ExitCode = &exitCode
		default:
			taskRun.State = taskrunner.JOB_RUN_STATE_UNKNOWN
		}
	} else {
		taskRun.State = taskrunner.JOB_RUN_STATE_SUCCESS
		exitCode := 0
		taskRun.ExitCode = &exitCode
	}
	now := taskrunner.Timestamp(providesNow())
	taskRun.EndTimestamp = &now

	if taskRunStatusChangeChan != nil {
		taskRunStatusChangeChan <- taskRun
	}

	return nil
}

func handleTaskrunnerError(errorMessage string, logFile io.Writer, jobRunStateChan chan *taskrunner.TaskRun, jobRun *taskrunner.TaskRun, providesNow NowProvider) errorsx.Error {
	now := taskrunner.Timestamp(providesNow())
	jobRun.EndTimestamp = &now
	jobRun.State = taskrunner.JOB_RUN_STATE_FAILED
	if jobRunStateChan != nil {
		jobRunStateChan <- jobRun
	}
	err := writeStringToLogFile(errorMessage, logFile, SourceTaskmasterHarness, providesNow)
	if nil != err {
		return errorsx.Wrap(err)
	}
	return nil
}
