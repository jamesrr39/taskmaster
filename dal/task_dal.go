package dal

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jamesrr39/taskmaster/domain"
	"gopkg.in/yaml.v2"
)

type TaskDAL struct {
	basePath string
}

func NewTaskDAL(basePath string) *TaskDAL {
	return &TaskDAL{basePath}
}

func (d *TaskDAL) GetAll() ([]*domain.Task, errorsx.Error) {
	tasksDirPath := filepath.Join(d.basePath, "tasks")
	entries, err := os.ReadDir(tasksDirPath)
	if err != nil {
		return nil, errorsx.Wrap(err, "tasksDirPath", tasksDirPath)
	}

	tasks := []*domain.Task{}
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

func readTaskFile(taskFilePath string) (*domain.Task, errorsx.Error) {
	f, err := os.Open(taskFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
	}
	defer f.Close()

	task := new(domain.Task)
	err = yaml.NewDecoder(f).Decode(task)
	if err != nil {
		return nil, errorsx.Wrap(err, "taskFilePath", taskFilePath)
	}
	task.Name = strings.TrimSuffix(filepath.Base(taskFilePath), ".yml")
	return task, nil
}
