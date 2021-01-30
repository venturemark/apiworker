package handler

import "github.com/xh3b4sd/rescue/pkg/task"

type Interface interface {
	Ensure(tsk *task.Task) error
	Filter(tsk *task.Task) bool
}
