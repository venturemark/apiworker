package timeline

import (
	"strconv"

	"github.com/venturemark/apicommon/pkg/metadata"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/rescue/pkg/task"
	"github.com/xh3b4sd/tracer"
)

type HandlerConfig struct {
	Logger logger.Interface
	Redigo redigo.Interface
}

type Handler struct {
	logger logger.Interface
	redigo redigo.Interface
}

func NewHandler(c HandlerConfig) (*Handler, error) {
	if c.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.Redigo == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Redigo must not be empty", c)
	}

	h := &Handler{
		logger: c.Logger,
		redigo: c.Redigo,
	}

	return h, nil
}

func (h *Handler) Ensure(tsk *task.Task) error {
	var err error

	var oid float64
	{
		oid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.OrganizationID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var tid float64
	{
		tid, err = strconv.ParseFloat(tsk.Obj.Metadata[metadata.TimelineID], 64)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	// delete message
	// delete update
	// delete timeline

	return nil
}

func (h *Handler) Filter(tsk *task.Task) bool {
	met := map[string]string{
		metadata.TaskAction:   "delete",
		metadata.TaskResource: "timeline",
	}

	return metadata.Contains(tsk.Obj.Metadata, met)
}
