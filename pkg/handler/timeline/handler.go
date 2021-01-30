package timeline

import (
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue/pkg/task"
)

type HandlerConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
}

type Handler struct{}

func NewHandler(c HandlerConfig) (*Handler, error) {
	return nil, nil
}

func (h *Handler) Ensure(tsk *task.Task) error {
	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	return false
}
