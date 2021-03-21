package subjectdelete

import (
	"context"
	"time"

	"github.com/venturemark/apicommon/pkg/key"
	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

var (
	resource = []string{
		"invite",
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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting subject associations")

	err = h.deleteSubject(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted subject associations")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	for _, r := range resource {
		met := map[string]string{
			metadata.TaskAction:   "delete",
			metadata.TaskResource: r,
		}

		if metadata.Contains(tsk.Obj.Metadata, met) {
			return true
		}
	}

	return false
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

func metaWithKind(met map[string]string, kind string) map[string]string {
	cop := map[string]string{}

	for k, v := range met {
		cop[k] = v
	}

	cop[metadata.ResourceKind] = kind

	return cop
}
