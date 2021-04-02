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
	"github.com/venturemark/apiworker/pkg/handler/invitedelete"
	"github.com/venturemark/apiworker/pkg/handler/messagedelete"
	"github.com/venturemark/apiworker/pkg/handler/roledelete"
	"github.com/venturemark/apiworker/pkg/handler/subjectdelete"
	"github.com/venturemark/apiworker/pkg/handler/timelinedelete"
	"github.com/venturemark/apiworker/pkg/handler/updatedelete"
	"github.com/venturemark/apiworker/pkg/handler/userdelete"
	"github.com/venturemark/apiworker/pkg/handler/venturedelete"
	"github.com/venturemark/apiworker/pkg/server"
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

	var inviteDeleteHandler handler.Interface
	{
		c := invitedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		inviteDeleteHandler, err = invitedelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var messageDeleteHandler handler.Interface
	{
		c := messagedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		messageDeleteHandler, err = messagedelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var roleDeleteHandler handler.Interface
	{
		c := roledelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		roleDeleteHandler, err = roledelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var subjectDeleteHandler handler.Interface
	{
		c := subjectdelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		subjectDeleteHandler, err = subjectdelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var timelineDeleteHandler handler.Interface
	{
		c := timelinedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		timelineDeleteHandler, err = timelinedelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var updateDeleteHandler handler.Interface
	{
		c := updatedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		updateDeleteHandler, err = updatedelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var userDeleteHandler handler.Interface
	{
		c := userdelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		userDeleteHandler, err = userdelete.NewHandler(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var ventureDeleteHandler handler.Interface
	{
		c := venturedelete.HandlerConfig{
			Logger: r.logger,
			Redigo: redigoClient,
			Rescue: rescueEngine,

			Timeout: r.flag.Handler.Timeout,
		}

		ventureDeleteHandler, err = venturedelete.NewHandler(c)
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
				inviteDeleteHandler,
				messageDeleteHandler,
				roleDeleteHandler,
				subjectDeleteHandler,
				timelineDeleteHandler,
				updateDeleteHandler,
				userDeleteHandler,
				ventureDeleteHandler,
			},
			Logger: r.logger,
			Rescue: rescueEngine,

			Interval: r.flag.Controller.Interval,
		}

		newController, err = queue.NewController(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var newServer *server.Server
	{
		c := server.Config{
			Logger: r.logger,

			ErrCha:   errCha,
			HTTPHost: r.flag.Metrics.Host,
			HTTPPort: r.flag.Metrics.Port,
		}

		newServer, err = server.New(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	{
		go newController.Boot()
		go newServer.ListenHTTP()
	}

	{
		signal.Notify(sigCha, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-errCha:
			return tracer.Mask(err)

		case <-sigCha:
			select {
			case <-time.After(r.flag.ApiWorker.TerminationGracePeriod):
			case <-sigCha:
			}

			return nil
		}
	}
}
