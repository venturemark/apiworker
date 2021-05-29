package daemon

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xh3b4sd/tracer"
)

type flag struct {
	ApiWorker struct {
		Host                   string
		Port                   string
		TerminationGracePeriod time.Duration
	}
	Controller struct {
		Interval time.Duration
	}
	Handler struct {
		Timeout time.Duration
	}
	Metrics struct {
		Host string
		Port string
	}
	Postmark struct {
		Token struct {
			Account string
			Server  string
		}
	}
	Redis struct {
		Host string
		Kind string
		Port string
	}
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.ApiWorker.Host, "apiworker-host", "", "127.0.0.1", "The host for binding the grpc apiworker to.")
	cmd.Flags().StringVarP(&f.ApiWorker.Port, "apiworker-port", "", "7777", "The port for binding the grpc apiworker to.")
	cmd.Flags().DurationVarP(&f.ApiWorker.TerminationGracePeriod, "apiworker-termination-grace-period", "", 5*time.Second, "The time to wait before terminating the apiworker process.")

	cmd.Flags().DurationVarP(&f.Controller.Interval, "controller-interval", "", 5*time.Second, "The interval of the controller to reconcile.")

	cmd.Flags().DurationVarP(&f.Handler.Timeout, "handler-timeout", "", 5*time.Second, "The timeout for a handler to give up.")

	cmd.Flags().StringVarP(&f.Metrics.Host, "metrics-host", "", "127.0.0.1", "The host for binding the http metrics endpoints to.")
	cmd.Flags().StringVarP(&f.Metrics.Port, "metrics-port", "", "8000", "The port for binding the http metrics endpoints to.")

	cmd.Flags().StringVarP(&f.Postmark.Token.Account, "postmark-token-account", "", os.Getenv("APIWORKER_POSTMARK_TOKEN_ACCOUNT"), "The postmark account token used to send emails.")
	cmd.Flags().StringVarP(&f.Postmark.Token.Server, "postmark-token-server", "", os.Getenv("APIWORKER_POSTMARK_TOKEN_SERVER"), "The postmark server token used to send emails.")

	cmd.Flags().StringVarP(&f.Redis.Host, "redis-host", "", "127.0.0.1", "The host for connecting with redis.")
	cmd.Flags().StringVarP(&f.Redis.Kind, "redis-kind", "", "single", "The kind of redis to connect to, e.g. simple or sentinel.")
	cmd.Flags().StringVarP(&f.Redis.Port, "redis-port", "", "6379", "The port for connecting with redis.")
}

func (f *flag) Validate() error {
	{
		if f.ApiWorker.Host == "" {
			return tracer.Maskf(invalidFlagError, "--apiworker-host must not be empty")
		}
		if f.ApiWorker.Port == "" {
			return tracer.Maskf(invalidFlagError, "--apiworker-port must not be empty")
		}
		if f.ApiWorker.TerminationGracePeriod == 0 {
			return tracer.Maskf(invalidFlagError, "--apiworker-termination-grace-period must not be empty")
		}
	}

	{
		if f.Controller.Interval == 0 {
			return tracer.Maskf(invalidFlagError, "--controller-interval must not be empty")
		}
	}

	{
		if f.Handler.Timeout == 0 {
			return tracer.Maskf(invalidFlagError, "--handler-timeout must not be empty")
		}
	}

	{
		if f.Metrics.Host == "" {
			return tracer.Maskf(invalidFlagError, "--metrics-host must not be empty")
		}
		if f.Metrics.Port == "" {
			return tracer.Maskf(invalidFlagError, "--metrics-port must not be empty")
		}
	}

	{
		if f.Postmark.Token.Account == "" {
			return tracer.Maskf(invalidFlagError, "--postmark-token-account must not be empty")
		}
		if f.Postmark.Token.Server == "" {
			return tracer.Maskf(invalidFlagError, "--postmark-token-server must not be empty")
		}
	}

	{
		if f.Redis.Host == "" {
			return tracer.Maskf(invalidFlagError, "--redis-host must not be empty")
		}
		if f.Redis.Kind == "" {
			return tracer.Maskf(invalidFlagError, "--redis-kind must not be empty")
		}
		if f.Redis.Port == "" {
			return tracer.Maskf(invalidFlagError, "--redis-port must not be empty")
		}
	}

	return nil
}
