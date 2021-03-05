package daemon

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/redigo/pkg/client"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/engine"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/pkg/controller"
	"github.com/venturemark/apiworker/pkg/controller/queue"
	"github.com/venturemark/apiworker/pkg/handler"
	"github.com/venturemark/apiworker/pkg/handler/audience"
	"github.com/venturemark/apiworker/pkg/handler/messagedelete"
	"github.com/venturemark/apiworker/pkg/handler/timeline"
	"github.com/venturemark/apiworker/pkg/handler/update"
)

type runner struct {
	flag   *flag
	logger logger.Interface
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return tracer.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return tracer.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	var redigoClient redigo.Interface
	{
		c := client.Config{
			Address: net.JoinHostPort(r.flag.Redis.Host, r.flag.Redis.Port),
			Kind:    r.flag.Redis.Kind,
		}

		redigoClient, err = client.New(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var rescueEngine rescue.Interface
	{
		c := engine.Config{
			Logger: r.logger,
			Redigo: redigoClient,
		}

		rescueEngine, err = engine.New(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var audienceHandler handler.Interface
	{
		c := audience.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,

			Timeout: r.flag.handler.Timeout,
		}

		audienceHandler, err = audience.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var messageDeleteHandler handler.Interface
	{
		c := messagedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,

			Timeout: r.flag.handler.Timeout,
		}

		messageDeleteHandler, err = messagedelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var timelineHandler handler.Interface
	{
		c := timeline.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,

			Timeout: r.flag.handler.Timeout,
		}

		timelineHandler, err = timeline.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var updateHandler handler.Interface
	{
		c := update.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.handler.Timeout,
		}

		updateHandler, err = update.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var donCha chan struct{}
	var errCha chan error
	var sigCha chan os.Signal
	{
		donCha = make(chan struct{})
		errCha = make(chan error, 1)
		sigCha = make(chan os.Signal, 2)

		defer close(donCha)
		defer close(errCha)
		defer close(sigCha)
	}

	var newController controller.Interface
	{
		c := queue.ControllerConfig{
			DonCha: donCha,
			ErrCha: errCha,
			Handler: []handler.Interface{
				audienceHandler,
				messageDeleteHandler,
				timelineHandler,
				updateHandler,
			},
			Logger: r.logger,
			Rescue: rescueEngine,

			Interval: r.flag.controller.Interval,
		}

		newController, err = queue.NewController(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		go newController.Boot()
	}

	{
		signal.Notify(sigCha, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-errCha:
			return tracer.Mask(err)

		case <-sigCha:
			select {
			case <-time.After(r.flag.apiworker.TerminationGracePeriod):
			case <-sigCha:
			}

			return nil
		}
	}
}
