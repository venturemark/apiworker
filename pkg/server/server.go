package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xh3b4sd/logger"
	"github.com/xh3b4sd/tracer"
)

type Config struct {
	Collector []prometheus.Collector
	Logger    logger.Interface

	ErrCha   chan<- error
	HTTPHost string
	HTTPPort string
}

type Server struct {
	collector []prometheus.Collector
	logger    logger.Interface

	errCha   chan<- error
	httpHost string
	httpPort string
}

func New(config Config) (*Server, error) {
	if len(config.Collector) == 0 {
		return nil, tracer.Maskf(invalidConfigError, "%T.Collector must not be empty", config)
	}
	if config.Logger == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ErrCha == nil {
		return nil, tracer.Maskf(invalidConfigError, "%T.ErrCha must not be empty", config)
	}
	if config.HTTPHost == "" {
		return nil, tracer.Maskf(invalidConfigError, "%T.HTTPHost must not be empty", config)
	}
	if config.HTTPPort == "" {
		return nil, tracer.Maskf(invalidConfigError, "%T.HTTPPort must not be empty", config)
	}

	s := &Server{
		collector: config.Collector,
		logger:    config.Logger,

		errCha:   config.ErrCha,
		httpHost: config.HTTPHost,
		httpPort: config.HTTPPort,
	}

	return s, nil
}

func (s *Server) ListenHTTP() {
	a := net.JoinHostPort(s.httpHost, s.httpPort)
	r := prometheus.NewPedanticRegistry()

	{
		for _, c := range s.collector {
			r.MustRegister(c)
		}
	}

	{
		http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	}

	s.logger.Log(context.Background(), "level", "info", "message", fmt.Sprintf("http server running at %s", a))

	{
		err := http.ListenAndServe(a, nil)
		if err != nil {
			s.errCha <- tracer.Mask(err)
		}
	}
}
