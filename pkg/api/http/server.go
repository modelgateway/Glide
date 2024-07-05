package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EinStack/glide/pkg/routers/manager"

	"github.com/gofiber/contrib/otelfiber"

	"github.com/gofiber/swagger"

	"github.com/EinStack/glide/docs"

	_ "github.com/EinStack/glide/docs" // importing docs package to include them into the binary
	"github.com/gofiber/contrib/fiberzap/v2"

	"github.com/gofiber/fiber/v2"

	"github.com/EinStack/glide/pkg/telemetry"
)

type Server struct {
	config        *ServerConfig
	telemetry     *telemetry.Telemetry
	routerManager *manager.RouterManager
	server        *fiber.App
}

func NewServer(config *ServerConfig, tel *telemetry.Telemetry, routerManager *manager.RouterManager) (*Server, error) {
	srv := config.ToServer()

	return &Server{
		config:        config,
		telemetry:     tel,
		routerManager: routerManager,
		server:        srv,
	}, nil
}

func (srv *Server) Run() error {
	// TODO: refactor this when https://github.com/gofiber/contrib/pull/1069 is merged
	srv.server.Get("/swagger.json", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).Type("json").Send(docs.SwaggerJSON)
	})

	srv.server.Use(otelfiber.Middleware())

	srv.server.Use(fiberzap.New(fiberzap.Config{
		Logger: srv.telemetry.Logger,
	}))

	v1 := srv.server.Group("/v1")

	v1.Get("/swagger/*", swagger.New(swagger.Config{
		Title: "Glide API Docs",
		URL:   "/swagger.json",
	}))

	v1.Get("/language/", LangRoutersHandler(srv.routerManager))
	v1.Post("/language/:router/chat/", LangChatHandler(srv.routerManager))

	v1.Use("/language/:router/chatStream", LangStreamRouterValidator(srv.routerManager))
	v1.Get("/language/:router/chatStream", LangStreamChatHandler(srv.telemetry, srv.routerManager))

	v1.Get("/health/", HealthHandler)

	srv.server.Use(NotFoundHandler)

	return srv.server.Listen(srv.config.Address())
}

func (srv *Server) Shutdown(ctx context.Context) error {
	exitWaitTime := 5 * time.Second

	srv.telemetry.Logger.Info(
		fmt.Sprintf("Begin graceful shutdown, wait at most %d seconds...", exitWaitTime/time.Second),
	)

	c, cancel := context.WithTimeout(ctx, exitWaitTime)
	defer cancel()

	if err := srv.server.ShutdownWithContext(c); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			srv.telemetry.Logger.Info("Server closed forcefully due to shutdown timeout")
			return nil
		}

		return err
	}

	return nil
}
