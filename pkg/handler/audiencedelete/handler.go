package audiencedelete

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/venturemark/apicommon/pkg/key"
	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

type HandlerConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
	Rescue rescue.Interface

	Timeout time.Duration
}

type Handler struct {
	logger logger.Interface
	redigo redigo.Interface
	rescue rescue.Interface

	timeout time.Duration
}

func NewHandler(c HandlerConfig) (*Handler, error) {
	if c.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.Redigo == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Redigo must not be empty", c)
	}
	if c.Rescue == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Rescue must not be empty", c)
	}

	if c.Timeout == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Timeout must not be empty", c)
	}

	h := &Handler{
		logger: c.Logger,
		redigo: c.Redigo,
		rescue: c.Rescue,

		timeout: c.Timeout,
	}

	return h, nil
}

func (h *Handler) Ensure(tsk *task.Task) error {
	var err error

	h.logger.Log(context.Background(), "level", "info", "message", "deleting audience resource")

	err = h.deleteElement(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted audience resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "audience",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteElement(tsk *task.Task) error {
	var err error

	var aui float64
	{
		aui, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.AudienceID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		k := fmt.Sprintf(key.Audience)
		s := aui

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
