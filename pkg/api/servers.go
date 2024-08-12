package api

import (
	"context"
	"sync"

	"github.com/EinStack/glide/pkg/router"

	"go.uber.org/zap"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/api/http"
)

type ServerManager struct {
	httpServer *http.Server
	shutdownWG *sync.WaitGroup
	telemetry  *telemetry.Telemetry
}

func NewServerManager(cfg *Config, tel *telemetry.Telemetry, router *router.Manager) (*ServerManager, error) {
	httpServer, err := http.NewServer(cfg.HTTP, tel, router)
	if err != nil {
		return nil, err
	}

	// TODO: init other servers like gRPC in future

	return &ServerManager{
		httpServer: httpServer,
		shutdownWG: &sync.WaitGroup{},
		telemetry:  tel,
	}, nil
}

func (mgr *ServerManager) Start() {
	if mgr.httpServer != nil {
		mgr.shutdownWG.Add(1)

		go func() {
			defer mgr.shutdownWG.Done()

			err := mgr.httpServer.Run()
			if err != nil {
				mgr.telemetry.Logger.Error("error on running HTTP server", zap.Error(err))
			}
		}()
	}
}

func (mgr *ServerManager) Shutdown(ctx context.Context) error {
	var err error

	if mgr.httpServer != nil {
		err = mgr.httpServer.Shutdown(ctx)
	}

	mgr.shutdownWG.Wait()

	return err
}
