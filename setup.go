package caddyprom

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	defaultPath = "/metrics"
	defaultAddr = "0.0.0.0:9180"
)

var (
	requestDuration *prometheus.SummaryVec
	responseSize    *prometheus.SummaryVec
)

func (m *Metrics) initMetrics(ctx caddy.Context) error {
	log := ctx.Logger(m)

	m.registerMetrics("caddy", "http")
	if m.Path == "" {
		m.Path = defaultPath
	}
	if m.Addr == "" {
		m.Addr = defaultAddr
	}

	if !m.useCaddyAddr {
		mux := http.NewServeMux()
		mux.Handle(m.Path, m.metricsHandler)

		srv := &http.Server{Handler: mux}
		// if m.Addr does not have a port just add the default one
		if !strings.Contains(m.Addr, ":") {
			m.Addr += ":" + strings.Split(defaultAddr, ":")[1]
		}
		zap.S().Info("Binding prometheus exporter to %s", m.Addr)
		listener, err := net.Listen("tcp", m.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen to %s: %w", m.Addr, err)
		}

		go func() {
			err := srv.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				log.Error("metrics handler's server failed to serve", zap.Error(err))
			}
		}()
	}
	return nil
}

func (m *Metrics) registerMetrics(namespace, subsystem string) {
	httpLabels := []string{"code", "method", "path"}

	requestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request took.",
	}, httpLabels)

	responseSize = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "response_size_bytes",
		Help:      "Size of the returns response in bytes.",
	}, httpLabels)
}
