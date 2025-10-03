package taskrunner

import (
	"errors"
)

type Script string

type Task struct {
	Id          uint      `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Script      Script    `json:"script"`
	Log         LogConfig `json:"log" required:"true"`
}

func NewTask(id uint, name string, description string, script Script) (*Task, error) {
	if name == "" {
		return nil, errors.New("A task must have a name")
	}

	return &Task{Id: id, Name: name, Description: description, Script: script}, nil
}
