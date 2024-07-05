package embed

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/resiliency/retry"
	"github.com/EinStack/glide/pkg/routers/lang"
	"github.com/EinStack/glide/pkg/telemetry"
	"go.uber.org/zap"
)

type EmbeddingRouter struct {
	routerID lang.RouterID
	Config   *LangRouterConfig
	retry    *retry.ExpRetry
	tel      *telemetry.Telemetry
	logger   *zap.Logger
}

func (r *EmbeddingRouter) Embed(ctx context.Context, req *schemas.EmbedRequest) (*schemas.EmbedResponse, error) {
}
