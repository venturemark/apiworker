package version

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"

	"github.com/venturemark/apiworker/pkg/project"
)

type runner struct {
	logger logger.Interface
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.run(ctx, cmd, args)
	if err != nil {
		return tracer.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	fmt.Fprintf(os.Stdout, "Git Commit    %s\n", project.GitSHA())
	fmt.Fprintf(os.Stdout, "Go Version    %s\n", runtime.Version())
	fmt.Fprintf(os.Stdout, "Go Arch       %s\n", runtime.GOARCH)
	fmt.Fprintf(os.Stdout, "Go OS         %s\n", runtime.GOOS)
	fmt.Fprintf(os.Stdout, "Source        %s\n", project.Source())
	fmt.Fprintf(os.Stdout, "Version       %s\n", project.Version())

	return nil
}
