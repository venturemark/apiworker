package timeline

import (
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
)

type HandlerConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
}

type Handler struct{}

func NewHandler(c HandlerConfig) (*Handler, error) {
	return nil, nil
}
