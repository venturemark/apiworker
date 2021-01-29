package daemon

import (
	"github.com/spf13/cobra"
	"github.com/xh3b4sd/tracer"
)

type flag struct {
	apiworker struct {
		Host string
		Port string
	}
	Redis struct {
		Host string
		Kind string
		Port string
	}
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.apiworker.Host, "apiworker-host", "", "127.0.0.1", "The host for binding the grpc apiworker to.")
	cmd.Flags().StringVarP(&f.apiworker.Port, "apiworker-port", "", "7777", "The port for binding the grpc apiworker to.")
	cmd.Flags().StringVarP(&f.Redis.Host, "redis-host", "", "127.0.0.1", "The host for connecting with redis.")
	cmd.Flags().StringVarP(&f.Redis.Kind, "redis-kind", "", "single", "The kind of redis to connect to, e.g. simple or sentinel.")
	cmd.Flags().StringVarP(&f.Redis.Port, "redis-port", "", "6379", "The port for connecting with redis.")
}

func (f *flag) Validate() error {
	{
		if f.apiworker.Host == "" {
			return tracer.Maskf(invalidFlagError, "--apiworker-host must not be empty")
		}
		if f.apiworker.Port == "" {
			return tracer.Maskf(invalidFlagError, "--apiworker-port must not be empty")
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
