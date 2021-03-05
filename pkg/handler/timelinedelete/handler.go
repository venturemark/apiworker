package timelinedelete

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

	h.logger.Log(context.Background(), "level", "info", "message", "deleting timeline")

	err = h.deleteElement(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	err = h.deleteKeys(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	h.logger.Log(context.Background(), "level", "info", "message", "deleted timeline")

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "timeline",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (h *Handler) deleteElement(tsk *task.Task) error {
	var err error

	var tid float64
	{
		tid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.TimelineID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var vid string
	{
		vid = tsk.Obj.Metadata[metadata.VentureID]
	}

	{
		k := fmt.Sprintf(key.Timeline, vid)
		s := tid

		err = h.redigo.Sorted().Delete().Score(k, s)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (h *Handler) deleteKeys(tsk *task.Task) error {
	var err error

	var vid string
	{
		vid = tsk.Obj.Metadata[metadata.VentureID]
	}

	var tid string
	{
		tid = tsk.Obj.Metadata[metadata.TimelineID]
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
		for k := range res {
			err = h.redigo.Simple().Delete().Element(k)
			if err != nil {
				erc <- tracer.Mask(err)
			}
		}
	}()

	go func() {
		defer close(don)
		defer close(erc)
		defer close(res)

		k := fmt.Sprintf(key.Update, vid, tid)

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
