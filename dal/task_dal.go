package dal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/db"
	"github.com/jamesrr39/taskmaster/taskexecutor"
	"github.com/jamesrr39/taskmaster/taskrunner"
	"gopkg.in/yaml.v2"

	"github.com/klauspost/compress/zstd"
)

type TaskDAL struct {
	basePath    string
	nowProvider taskexecutor.NowProvider
}

func NewTaskDAL(basePath string, nowProvider taskexecutor.NowProvider) *TaskDAL {
	return &TaskDAL{basePath, nowProvider}
}

func (d *TaskDAL) GetAll() ([]*taskrunner.Task, errorsx.Error) {
	tasksDirPath := filepath.Join(d.basePath, "tasks")
	entries, err := os.ReadDir(tasksDirPath)
	if err != nil {
		return nil, errorsx.Wrap(err, "tasksDirPath", tasksDirPath)
	}

	tasks := []*taskrunner.Task{}
	for _, entry := range entries {
		taskFilePath := filepath.Join(tasksDirPath, entry.Name())
		task, err := readTaskFile(taskFilePath)
		if err != nil {
			return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (d *TaskDAL) GetByName(name string) (*taskrunner.Task, errorsx.Error) {
	taskFilePath := filepath.Join(d.basePath, "tasks", name+".yml")
	task, err := readTaskFile(taskFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
	}

	return task, nil
}

func (d *TaskDAL) createTaskRun(dbConn db.DBConn, task *taskrunner.Task) (*taskrunner.TaskRun, errorsx.Error) {
	startTimestamp := taskrunner.Timestamp(d.nowProvider())

	type responseType struct {
		RunNumber uint64 `db:"task_run_number"`
	}
	response := new(responseType)

	err := dbConn.Get(
		response,
		`INSERT INTO task_runs (task_name, task_run_number, start_time)
		VALUES (
			$1,
			(SELECT COALESCE(MAX(task_run_number), 0) +1 FROM task_runs WHERE task_name = $1),
			$2
		)
		RETURNING task_run_number`,
		task.Name, startTimestamp)
	if err != nil {
		return nil, errorsx.Wrap(err)
	}

	taskRun := task.NewTaskRun(response.RunNumber, startTimestamp)

	return taskRun, nil
}

func (d *TaskDAL) insertTaskRunResults(dbConn db.DBConn, taskRun *taskrunner.TaskRun) errorsx.Error {
	_, err := dbConn.Exec(
		`INSERT INTO task_runs_results (task_name, task_run_number, end_time, exit_code)
		VALUES ($1, $2, $3, $4);
		`,
		taskRun.TaskName, taskRun.RunNumber, taskRun.EndTimestamp, taskRun.ExitCode,
	)
	if err != nil {
		return errorsx.Wrap(err)
	}

	return nil
}

func (d *TaskDAL) GetTaskRun(dbConn db.DBConn, taskName string, taskRunNumber uint64) (*taskrunner.TaskRun, errorsx.Error) {
	taskRun := new(taskrunner.TaskRun)

	err := dbConn.Get(
		taskRun,
		`
		SELECT tr.task_name, tr.task_run_number, start_time, end_time, exit_code
		FROM task_runs tr
		LEFT JOIN task_runs_results trr
		ON tr.task_name = trr.task_name
		AND tr.task_run_number = trr.task_run_number
		WHERE tr.task_name = $1 AND tr.task_run_number = $2;
		`,
		taskName, taskRunNumber,
	)
	if err != nil {
		return nil, errorsx.Wrap(err)
	}

	return taskRun, nil
}

func (d *TaskDAL) GetLogsTask(taskName string, runNumber uint64) (io.ReadCloser, errorsx.Error) {
	filePath := filepath.Join(d.basePath, "data", "results", taskName, "runs", fmt.Sprintf("%d", runNumber), "logs.jsonl.zst")

	f, err := os.Open(filePath)
	if err != nil {
		return nil, errorsx.Wrap(err, "filePath", filePath)
	}

	decoder, err := zstd.NewReader(f)
	if err != nil {
		return nil, errorsx.Wrap(err, "filePath", filePath)
	}

	return zstdReadCloser{f, decoder}, nil
}

type zstdReadCloser struct {
	file        io.ReadCloser
	zstdDecoder *zstd.Decoder
}

func (z zstdReadCloser) Read(b []byte) (int, error) {
	return z.zstdDecoder.Read(b)
}

func (z zstdReadCloser) Close() error {
	z.zstdDecoder.Close()
	return z.file.Close()
}

func (d *TaskDAL) RunTask(dbConn db.DBConn, task *taskrunner.Task) (*taskrunner.TaskRun, errorsx.Error) {
	var err error

	taskRun, err := d.createTaskRun(dbConn, task)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	taskRunTempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	taskRunDir := filepath.Join(d.basePath, "data", "results", task.Name, "runs", fmt.Sprintf("%d", taskRun.RunNumber))
	err = os.MkdirAll(taskRunDir, 0755)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun, "taskRunDir", taskRunDir)
	}

	logFilePath := filepath.Join(taskRunDir, "logs.jsonl.zst")

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}
	defer logFile.Close()

	zstdWriter, err := zstd.NewWriter(logFile)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}
	defer func() {
		err := zstdWriter.Flush()
		if err != nil {
			slog.Error("couldn't flush zstd writer", "taskRun", taskRun, "error", err)
		}

		err = zstdWriter.Close()
		if err != nil {
			slog.Error("couldn't close zstd writer", "taskRun", taskRun, "error", err)
		}
	}()

	err = taskexecutor.ExecuteJobRun(task, taskRun, nil, zstdWriter, taskRunTempDir, d.nowProvider)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	err = d.insertTaskRunResults(dbConn, taskRun)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	return taskRun, nil
}

func readTaskFile(taskFilePath string) (*taskrunner.Task, errorsx.Error) {
	f, err := os.Open(taskFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
	}
	defer f.Close()

	task := new(taskrunner.Task)
	err = yaml.NewDecoder(f).Decode(task)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
	}
	task.Name = strings.TrimSuffix(filepath.Base(taskFilePath), ".yml")
	return task, nil
}
