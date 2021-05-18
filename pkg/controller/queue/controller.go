package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/mutant"
	"github.com/xh3b4sd/mutant/pkg/wave"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/engine"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/pkg/handler"
)

type ControllerConfig struct {
	DonCha  <-chan struct{}
	ErrCha  chan<- error
	Handler []handler.Interface
	Logger  logger.Interface
	Redigo  redigo.Interface
	Rescue  rescue.Interface

	Interval time.Duration
}

type Controller struct {
	donCha  <-chan struct{}
	errCha  chan<- error
	handler []handler.Interface
	logger  logger.Interface
	redigo  redigo.Interface
	rescue  rescue.Interface

	mutant mutant.Interface

	interval time.Duration
}

func NewController(config ControllerConfig) (*Controller, error) {
	if config.DonCha == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.DonCha must not be empty", config)
	}
	if config.ErrCha == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.ErrCha must not be empty", config)
	}
	if len(config.Handler) == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Handler must not be empty", config)
	}
	if config.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Redigo == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Redigo must not be empty", config)
	}
	if config.Rescue == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Rescue must not be empty", config)
	}

	if config.Interval == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Interval must not be empty", config)
	}

	var err error

	var m mutant.Interface
	{
		c := wave.Config{
			Length: 2,
		}

		m, err = wave.New(c)
		if err != nil {
			return nil, tracer.Mask(err)
		}
	}

	c := &Controller{
		donCha:  config.DonCha,
		errCha:  config.ErrCha,
		handler: config.Handler,
		logger:  config.Logger,
		redigo:  config.Redigo,
		rescue:  config.Rescue,

		mutant: m,

		interval: config.Interval,
	}

	return c, nil
}

func (c *Controller) Boot() {
	var err error

	c.logger.Log(context.Background(), "level", "info", "message", fmt.Sprintf("controller reconciling every %s", c.interval.String()))

	for {
		select {
		case <-c.donCha:
			return
		case <-time.After(c.interval):
			err = c.createTasks()
			if IsDialError(err) {
				c.logger.Log(context.Background(), "level", "warning", "message", "connection refused")
			} else if err != nil {
				c.errCha <- tracer.Mask(err)
			}

			err = c.searchTasks()
			if IsDialError(err) {
				c.logger.Log(context.Background(), "level", "warning", "message", "connection refused")
			} else if err != nil {
				c.errCha <- tracer.Mask(err)
			}
		}
	}
}

func (c *Controller) createTasks() error {
	o := func() error {
		t := &task.Task{
			Obj: task.TaskObj{
				Metadata: map[string]string{
					metadata.TaskAction:   "create",
					metadata.TaskInterval: "weekly",
					metadata.TaskResource: "reminder",
				},
			},
		}

		err := c.rescue.Create(t)
		if err != nil {
			return tracer.Mask(err)
		}

		return nil
	}

	{
		err := c.weekly(o)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}

func (c *Controller) searchTasks() error {
	{
		select {
		case <-c.mutant.Check():
			c.mutant.Reset()
		default:
		}

		l := c.mutant.Index()

		if l[0] == 0 && l[1] == 0 {
			c.mutant.Shift()
		}

		defer c.mutant.Shift()
	}

	{
		l := c.mutant.Index()

		if l[0] == 1 {
			err := c.rescue.Expire()
			if err != nil {
				return tracer.Mask(err)
			}
		}

		if l[1] == 1 {
			tsk, err := c.rescue.Search()
			if engine.IsNoTask(err) {
				return nil
			} else if err != nil {
				return tracer.Mask(err)
			}

			c.logger.Log(context.Background(), "level", "info", "message", "reconciling task", "resource", tsk.Obj.Metadata[metadata.TaskResource])
			defer c.logger.Log(context.Background(), "level", "info", "message", "reconciled task", "resource", tsk.Obj.Metadata[metadata.TaskResource])

			var inc bool
			for _, h := range c.handler {
				if h.Filter(tsk) {
					err = h.Ensure(tsk)
					if IsIncompleteExecution(err) {
						inc = true
					} else if err != nil {
						return tracer.Mask(err)
					}
				}
			}

			if inc {
				// Upon incomplete task execution we just move on to the next
				// task without deleting the current task. The unfinished task
				// will time out and be rescheduled, causing some worker process
				// to pick it up eventually at a later point in time.
			} else {
				err = c.rescue.Delete(tsk)
				if err != nil {
					return tracer.Mask(err)
				}
			}
		}
	}

	return nil
}

func (c *Controller) weekly(o func() error) error {
	var t time.Time
	{
		t = time.Now().UTC()
		m := t.Minute()

		if t.Weekday() != time.Monday {
			return nil
		}

		if m != 0 {
			return nil
		}
	}

	var k string
	var v string
	{
		k = "apiworker.venturemark.co:rem:wee"
		v = fmt.Sprintf("%02d.%02d.%d", t.Day(), t.Month(), t.Year())
	}

	{
		val, err := c.redigo.Simple().Search().Value(k)
		if err != nil {
			return tracer.Mask(err)
		}

		if v == val {
			return nil
		}
	}

	{
		err := o()
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		err := c.redigo.Simple().Create().Element(k, v)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	return nil
}
