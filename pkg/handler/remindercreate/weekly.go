package remindercreate

import (
	"context"
	"strings"
	"time"

	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

type WeeklyConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
	Rescue rescue.Interface

	Timeout time.Duration
}

type Weekly struct {
	logger logger.Interface
	redigo redigo.Interface
	rescue rescue.Interface

	timeout time.Duration
}

func NewWeekly(c WeeklyConfig) (*Weekly, error) {
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

	w := &Weekly{
		logger: c.Logger,
		redigo: c.Redigo,
		rescue: c.Rescue,

		timeout: c.Timeout,
	}

	return w, nil
}

func (w *Weekly) Ensure(tsk *task.Task) error {
	var err error

	w.logger.Log(context.Background(), "level", "info", "message", "creating weekly reminder")

	err = w.createReminder(tsk)
	if err != nil {
		return tracer.Mask(err)
	}

	w.logger.Log(context.Background(), "level", "info", "message", "created weekly reminder")

	return nil
}

func (w *Weekly) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "create",
		metadata.TaskInterval: "weekly",
		metadata.TaskResource: "reminder",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}

func (w *Weekly) createReminder(tsk *task.Task) error {
	var err error

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
			var uid string
			{
				uid = strings.TrimPrefix(k, "use:")
			}

			t := &task.Task{
				Obj: task.TaskObj{
					Metadata: map[string]string{
						metadata.TaskAction:   "create",
						metadata.TaskAudience: "user",
						metadata.TaskResource: "reminder",

						"user.venturemark.co/id": uid,
					},
				},
			}

			err := w.rescue.Create(t)
			if err != nil {
				erc <- tracer.Mask(err)
			}
		}
	}()

	go func() {
		defer close(res)

		k := "use:[0-9]*[0-9][^:]"

		err = w.redigo.Walker().Simple(k, don, res)
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

		case <-time.After(w.timeout):
			return tracer.Mask(timeoutError)
		}
	}
}
