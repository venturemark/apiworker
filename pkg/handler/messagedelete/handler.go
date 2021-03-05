package messagedelete

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/venturemark/apicommon/pkg/key"
	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

type HandlerConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface

	Timeout time.Duration
}

type Handler struct {
	logger logger.Interface
	redigo redigo.Interface

	timeout time.Duration
}

func NewHandler(c HandlerConfig) (*Handler, error) {
	if c.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.Redigo == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Redigo must not be empty", c)
	}

	if c.Timeout == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Timeout must not be empty", c)
	}

	h := &Handler{
		logger: c.Logger,
		redigo: c.Redigo,

		timeout: c.Timeout,
	}

	return h, nil
}

func (h *Handler) Ensure(tsk *task.Task) error {
	var err error

	h.logger.Log(context.Background(), "level", "info", "message", "deleting message resource")

	err = h.deleteElement(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted message resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "message",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteElement(tsk *task.Task) error {
	var err error

	var mid float64
	{
		mid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.MessageID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var tid string
	{
		tid = tsk.Obj.Metadata[metadata.TimelineID]
	}

	var uid string
	{
		uid = tsk.Obj.Metadata[metadata.UpdateID]
	}

	var vid string
	{
		vid = tsk.Obj.Metadata[metadata.VentureID]
	}

	{
		k := fmt.Sprintf(key.Message, vid, tid, uid)
		s := mid

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
