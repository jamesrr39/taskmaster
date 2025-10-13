package taskrunner

import (
	"errors"
)

type Script string

type Task struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Script      Script `json:"script"`
}

func NewTask(name string, description string, script Script) (*Task, error) {
	if name == "" {
		return nil, errors.New("a task must have a name")
	}

	return &Task{Name: name, Description: description, Script: script}, nil
}
