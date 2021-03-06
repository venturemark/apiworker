package timelinedelete

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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting timeline resource")

	err = h.deleteTimeline(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteUpdate(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted timeline resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "timeline",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteTimeline(tsk *task.Task) error {
	var err error

	var tii float64
	{
		tii, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.TimelineID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var vei string
	{
		vei = tsk.Obj.Metadata[metadata.VentureID]
	}

	{
		k := fmt.Sprintf(key.Timeline, vei)
		s := tii

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deleteUpdate(tsk *task.Task) error {
	var vei string
	{
		vei = tsk.Obj.Metadata[metadata.VentureID]
	}

	var tii string
	{
		tii = tsk.Obj.Metadata[metadata.TimelineID]
	}

	var upd []*schema.Update
	{
		k := fmt.Sprintf(key.Update, vei, tii)
		str, err := h.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return tracer.Mask(err)
		}

		for _, s := range str {
			u := &schema.Update{}
			err = json.Unmarshal([]byte(s), u)
			if err != nil {
				return tracer.Mask(err)
			}

			upd = append(upd, u)
		}
	}

	for _, u := range upd {
		t := &task.Task{
			Obj: task.TaskObj{
				Metadata: u.Obj.Metadata,
			},
		}

		t.Obj.Metadata[metadata.TaskAction] = "delete"
		t.Obj.Metadata[metadata.TaskResource] = "update"

		err := h.rescue.Create(t)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
