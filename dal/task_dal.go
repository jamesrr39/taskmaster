package dal

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/taskexecutor"
	"github.com/jamesrr39/taskmaster/taskrunner"
	"gopkg.in/yaml.v2"
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

func (d *TaskDAL) createTaskRun(task *taskrunner.Task) (*taskrunner.TaskRun, errorsx.Error) {
	panic("not implemented")
}

func (d *TaskDAL) RunTask(task *taskrunner.Task) (*taskrunner.TaskRun, errorsx.Error) {
	var err error

	taskRun, err := d.createTaskRun(task)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, errorsx.Wrap(err, "taskRun", taskRun)
	}

	err = taskexecutor.ExecuteJobRun(taskRun, nil, os.Stderr, workDir, d.nowProvider)
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
