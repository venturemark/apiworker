package userdelete

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

var (
	resource = []string{
		"message",
		"timeline",
		"update",
		"user",
		"venture",
	}
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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting user resource")

	err = h.deleteAssociation(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteRole(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteSubject(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteUser(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted user resource")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "user",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteAssociation(tsk *task.Task) error {
	var err error

	var clk *key.Key
	{
		clk = key.Claim(tsk.Obj.Metadata)
	}

	{
		k := clk.Elem()

		err = h.redigo.Simple().Delete().Element(k)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
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

	{
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

func (h *Handler) deleteSubject(tsk *task.Task) error {
	var err error

	for _, r := range resource {
		met := metaWithKind(tsk.Obj.Metadata, r)

		var suk *key.Key
		{
			suk = key.Subject(met)
		}

		{
			k := suk.Elem()

			err = h.redigo.Sorted().Delete().Clean(k)
			if err != nil {
				return tracer.Mask(err)
			}
		}
	}

	return nil
}

func (h *Handler) deleteUser(tsk *task.Task) error {
	var err error

	var usk *key.Key
	{
		usk = key.User(tsk.Obj.Metadata)
	}

	{
		k := usk.Elem()

		err = h.redigo.Simple().Delete().Element(k)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func metaWithKind(met map[string]string, kind string) map[string]string {
	cop := map[string]string{}

	for k, v := range met {
		cop[k] = v
	}

	cop[metadata.ResourceKind] = kind

	return cop
}
