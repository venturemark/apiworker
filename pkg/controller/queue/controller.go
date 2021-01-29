package queue

import (
	"github.com/venturemark/apiworker/pkg/handler"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/rescue"
)

type ControllerConfig struct {
	Logger  logger.Interface
	Handler []handler.Interface
	Rescue  rescue.Interface
}

type Controller struct{}

func NewController(c ControllerConfig) (*Controller, error) {
	return nil, nil
}

func (c *Controller) Boot() error {
	return nil
}
