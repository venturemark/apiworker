package updatedelete

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/venturemark/apicommon/pkg/key"
	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/venturemark/apicommon/pkg/schema"
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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting update resource")

	err = h.deleteUpdate(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteMessage(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted update resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "update",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteUpdate(tsk *task.Task) error {
	var err error

	var tid string
	{
		tid = tsk.Obj.Metadata[metadata.TimelineID]
	}

	var uid float64
	{
		uid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.UpdateID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var vid string
	{
		vid = tsk.Obj.Metadata[metadata.VentureID]
	}

	{
		k := fmt.Sprintf(key.Update, vid, tid)
		s := uid

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deleteMessage(tsk *task.Task) error {
	var vid string
	{
		vid = tsk.Obj.Metadata[metadata.VentureID]
	}

	var tid string
	{
		tid = tsk.Obj.Metadata[metadata.TimelineID]
	}

	var uid string
	{
		uid = tsk.Obj.Metadata[metadata.UpdateID]
	}

	var mes []*schema.Message
	{
		k := fmt.Sprintf(key.Message, vid, tid, uid)
		str, err := h.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return tracer.Mask(err)
		}

		for _, s := range str {
			m := &schema.Message{}
			err = json.Unmarshal([]byte(s), m)
			if err != nil {
				return tracer.Mask(err)
			}

			mes = append(mes, m)
		}
	}

	for _, m := range mes {
		t := &task.Task{
			Obj: task.TaskObj{
				Metadata: m.Obj.Metadata,
			},
		}

		t.Obj.Metadata[metadata.TaskAction] = "delete"
		t.Obj.Metadata[metadata.TaskResource] = "message"

		err := h.rescue.Create(t)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
