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
	"github.com/xh3b4sd/mutant"
	"github.com/xh3b4sd/mutant/pkg/wave"
	"github.com/xh3b4sd/redigo"
	"github.com/xh3b4sd/redigo/pkg/client"
	"github.com/xh3b4sd/rescue"
	"github.com/xh3b4sd/rescue/pkg/engine"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/pkg/controller"
	"github.com/venturemark/apiworker/pkg/controller/queue"
	"github.com/venturemark/apiworker/pkg/handler"
	"github.com/venturemark/apiworker/pkg/handler/timeline"
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

	var timelineHandler handler.Interface
	{
		c := timeline.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
		}

		timelineHandler, err = timeline.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var waveMutant mutant.Interface
	{
		c := wave.Config{
			Length: 2,
		}

		waveMutant, err = wave.New(c)
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
	}

	var newController controller.Interface
	{
		c := queue.ControllerConfig{
			DonCha: donCha,
			ErrCha: errCha,
			Handler: []handler.Interface{
				timelineHandler,
			},
			Logger: r.logger,
			Mutant: waveMutant,
			Rescue: rescueEngine,
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
			close(donCha)
			defer close(errCha)
			defer close(sigCha)

			return tracer.Mask(err)

		case <-sigCha:
			close(donCha)
			defer close(errCha)
			defer close(sigCha)

			select {
			case <-time.After(5 * time.Second):
			case <-sigCha:
			}

			return nil
		}
	}
}
