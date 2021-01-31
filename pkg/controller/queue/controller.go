package queue

import (
	"time"

	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/mutant"
	"github.com/xh3b4sd/mutant/pkg/wave"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/engine"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/pkg/handler"
)

type ControllerConfig struct {
	DonCha  <-chan struct{}
	ErrCha  chan<- error
	Handler []handler.Interface
	Logger  logger.Interface
	Rescue  rescue.Interface
}

type Controller struct {
	donCha  <-chan struct{}
	errCha  chan<- error
	handler []handler.Interface
	logger  logger.Interface
	rescue  rescue.Interface

	mutant mutant.Interface
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
	if config.Rescue == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Rescue must not be empty", config)
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
		rescue:  config.Rescue,

		mutant: m,
	}

	return c, nil
}

func (c *Controller) Boot() {
	for {
		select {
		case <-c.donCha:
			return
		case <-time.After(5 * time.Second):
			err := c.bootE()
			if err != nil {
				c.errCha <- tracer.Mask(err)
			}
		}
	}
}

func (c *Controller) bootE() error {
	{
		l := c.mutant.Index()

		if l[0] == 0 && l[1] == 0 {
			c.mutant.Shift()
		}
	}

	{
		defer c.mutant.Shift()
	}

	select {
	case <-c.mutant.Check():
		c.mutant.Reset()
	default:
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

			for _, h := range c.handler {
				if h.Filter(tsk) {
					err = h.Ensure(tsk)
					if err != nil {
						return tracer.Mask(err)
					}

					err = c.rescue.Delete(tsk)
					if err != nil {
						return tracer.Mask(err)
					}
				}
			}
		}
	}

	return nil
}
