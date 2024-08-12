package pkg

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EinStack/glide/pkg/routers"

	"github.com/EinStack/glide/pkg/version"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"

	"github.com/EinStack/glide/pkg/config"

	"github.com/EinStack/glide/pkg/telemetry"

	"go.uber.org/zap"

	"github.com/EinStack/glide/pkg/api"

	"go.uber.org/multierr"
)

// Gateway represents an instance of running Glide gateway.
// It loads configs, start API server(s), and listen to termination signals to shut down
type Gateway struct {
	// configProvider holds all configurations
	configProvider *config.Provider
	// tel holds logger, meter, and tracer
	tel *telemetry.Telemetry
	// serverManager controls API over different protocols
	serverManager *api.ServerManager
	// signalChannel is used to receive termination signals from the OS.
	signalC chan os.Signal
	// shutdownC is used to terminate the gateway
	shutdownC chan struct{}
}

func NewGateway(configProvider *config.Provider) (*Gateway, error) {
	cfg := configProvider.Get()

	tel, err := telemetry.NewTelemetry(&telemetry.Config{LogConfig: cfg.Telemetry.LogConfig})
	if err != nil {
		return nil, err
	}

	tel.L().Info("🐦Glide is starting up", zap.String("version", version.FullVersion))
	tel.L().Debug("✅ Config loaded successfully:\n" + configProvider.GetStr())

	routerManager, err := routers.NewManager(&cfg.Routers, tel)
	if err != nil {
		return nil, err
	}

	serverManager, err := api.NewServerManager(cfg.API, tel, routerManager)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		configProvider: configProvider,
		tel:            tel,
		serverManager:  serverManager,
		signalC:        make(chan os.Signal, 3), // equal to number of signal types we expect to receive
		shutdownC:      make(chan struct{}),
	}, nil
}

// Run starts and runs the gateway according to given configuration
func (gw *Gateway) Run(ctx context.Context) error {
	// Instrument the gateway process
	err := host.Start()
	if err != nil {
		return err
	}

	err = runtime.Start()
	if err != nil {
		return err
	}

	gw.configProvider.Start()
	gw.serverManager.Start() //nolint:contextcheck

	signal.Notify(gw.signalC, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(gw.signalC)

LOOP:
	for {
		select {
		// TODO: Watch for config updates
		case sig := <-gw.signalC:
			gw.tel.L().Info("received signal from os", zap.String("signal", sig.String()))
			break LOOP
		case <-gw.shutdownC:
			gw.tel.L().Info("received shutdown request")
			break LOOP
		case <-ctx.Done():
			gw.tel.L().Info("context done, terminating process")
			// Call shutdown with background context as the passed in context has been canceled
			return gw.shutdown(context.Background()) //nolint:contextcheck
		}
	}

	return gw.shutdown(ctx)
}

func (gw *Gateway) Shutdown() {
	close(gw.shutdownC)
}

func (gw *Gateway) shutdown(ctx context.Context) error {
	var errs error

	if err := gw.serverManager.Shutdown(ctx); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to shutdown servers: %w", err))
	}

	return errs
}
