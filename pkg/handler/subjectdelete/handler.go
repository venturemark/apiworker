package subjectdelete

import (
	"context"
	"fmt"
	"time"

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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting subject associations")

	err = h.deleteSubject(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted subject associations")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "user",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteSubject(tsk *task.Task) error {
	var err error

	var sui string
	{
		sui = tsk.Obj.Metadata[metadata.SubjectID]
	}

	var don chan struct{}
	var erc chan error
	var res chan string
	{
		don = make(chan struct{}, 1)
		erc = make(chan error, 1)
		res = make(chan string, 1)
	}

	go func() {
		defer close(don)

		for k := range res {
			err = h.redigo.Sorted().Delete().Clean(k)
			if err != nil {
				erc <- tracer.Mask(err)
			}
		}
	}()

	go func() {
		defer close(res)

		k := fmt.Sprintf("*sub:%s*", sui)

		err = h.redigo.Walker().Simple(k, don, res)
		if err != nil {
			erc <- tracer.Mask(err)
		}
	}()

	{
		select {
		case <-don:
			return nil

		case err := <-erc:
			return tracer.Mask(err)

		case <-time.After(h.timeout):
			return tracer.Mask(timeoutError)
		}
	}
}
