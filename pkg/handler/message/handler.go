package message

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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting message")

	err = h.deleteElement(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deletePermission(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted message")

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

	var aid float64
	{
		aid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.AudienceID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var oid string
	{
		oid = tsk.Obj.Metadata[metadata.OrganizationID]
	}

	{
		k := fmt.Sprintf(key.Audience, oid)
		s := aid

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deletePermission(tsk *task.Task) error {
	var err error

	var oid string
	{
		oid = req.Obj.Metadata[metadata.OrganizationID]
	}

	var tid string
	{
		tid = req.Obj.Metadata[metadata.TimelineID]
	}

	var uid string
	{
		uid = req.Obj.Metadata[metadata.UpdateID]
	}

	{
		k := fmt.Sprintf(key.Owner, fmt.Sprintf(key.Message, oid, tid, uid))

		err = h.redigo.Simple().Delete().Element(k)
		if err != nil {
			return tracer.Mask(err)
		}
	}
	{
		k := fmt.Sprintf(key.Message, oid, tid, uid)
		s := mid

		err = c.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	return nil
}
