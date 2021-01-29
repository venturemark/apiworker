package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/cmd"
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		tracer.Panic(err)
	}
}

func mainE(ctx context.Context) error {
	var err error

	var l logger.Interface
	{
		c := logger.Config{}

		l, err = logger.New(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	var r *cobra.Command
	{
		c := cmd.Config{
			Logger: l,
		}

		r, err = cmd.New(c)
		if err != nil {
			return tracer.Mask(err)
		}
	}

	err = r.Execute()
	if err != nil {
		return tracer.Mask(err)
	}

	return nil
}
