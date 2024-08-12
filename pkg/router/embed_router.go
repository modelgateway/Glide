package router

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/telemetry"
)

type EmbedRouter struct {
	// routerID lang.RouterID
	// Config   *LangRouterConfig
	// retry  *retry.ExpRetry
	// tel    *telemetry.Telemetry
	// logger *zap.Logger
}

func NewEmbedRouter(_ *EmbedRouterConfig, _ *telemetry.Telemetry) (*EmbedRouter, error) {
	// TODO: implement
	return &EmbedRouter{}, nil
}

func (r *EmbedRouter) Embed(ctx context.Context, req *schema.EmbedRequest) (*schema.EmbedResponse, error) {
	// TODO: implement
}
