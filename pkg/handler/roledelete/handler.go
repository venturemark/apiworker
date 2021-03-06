package roledelete

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/venturemark/apicommon/pkg/hash"
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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting role resource")

	err = h.deleteElement(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted role resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "role",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteElement(tsk *task.Task) error {
	var err error

	var rei string
	{
		switch tsk.Obj.Metadata[metadata.ResourceKind] {
		case "audience":
			rei = hash.Audience(tsk.Obj.Metadata)
		case "message":
			rei = hash.Message(tsk.Obj.Metadata)
		case "timeline":
			rei = hash.Timeline(tsk.Obj.Metadata)
		case "update":
			rei = hash.Update(tsk.Obj.Metadata)
		case "venture":
			rei = hash.Venture(tsk.Obj.Metadata)
		}
	}

	var roi float64
	{
		roi, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.RoleID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		k := fmt.Sprintf(key.Role, rei)
		s := roi

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
