package venturedelete

import (
	"context"
	"encoding/json"
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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting venture resource")

	err = h.deleteRole(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteTimeline(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteVenture(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted venture resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "venture",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteRole(tsk *task.Task) error {
	var err error

	var rok *key.Key
	{
		rok = key.Role(tsk.Obj.Metadata)
	}

	var usi string
	{
		usi = tsk.Obj.Metadata[metadata.UserID]
	}

	var rol *schema.Role
	{
		k := rok.List()
		s := usi

		str, err := h.redigo.Sorted().Search().Index(k, s)
		if err != nil {
			return tracer.Mask(err)
		}

		if str != "" {
			rol = &schema.Role{}
			err = json.Unmarshal([]byte(str), rol)
			if err != nil {
				return tracer.Mask(err)
			}
		}
	}

	if rol != nil {
		t := &task.Task{
			Obj: task.TaskObj{
				Metadata: rol.Obj.Metadata,
			},
		}

		t.Obj.Metadata[metadata.TaskAction] = "delete"
		t.Obj.Metadata[metadata.TaskResource] = "role"

		err := h.rescue.Create(t)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		k := rok.List()

		err = h.redigo.Sorted().Delete().Clean(k)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deleteTimeline(tsk *task.Task) error {
	var tik *key.Key
	{
		tik = key.Timeline(tsk.Obj.Metadata)
	}

	var lis []*schema.Timeline
	{
		k := tik.List()

		str, err := h.redigo.Sorted().Search().Order(k, 0, -1)
		if err != nil {
			return tracer.Mask(err)
		}

		for _, s := range str {
			t := &schema.Timeline{}
			err = json.Unmarshal([]byte(s), t)
			if err != nil {
				return tracer.Mask(err)
			}

			lis = append(lis, t)
		}
	}

	for _, l := range lis {
		t := &task.Task{
			Obj: task.TaskObj{
				Metadata: l.Obj.Metadata,
			},
		}

		t.Obj.Metadata[metadata.TaskAction] = "delete"
		t.Obj.Metadata[metadata.TaskResource] = "timeline"

		err := h.rescue.Create(t)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deleteVenture(tsk *task.Task) error {
	var err error

	var vek *key.Key
	{
		vek = key.Venture(tsk.Obj.Metadata)
	}

	{
		k := vek.Elem()

		err = h.redigo.Sorted().Delete().Clean(k)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
