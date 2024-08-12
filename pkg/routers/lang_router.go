package routers

import (
	"context"
	"errors"

	"github.com/EinStack/glide/pkg/extmodel"

	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/resiliency/retry"
	"github.com/EinStack/glide/pkg/routers/routing"
	"github.com/EinStack/glide/pkg/telemetry"
	"go.uber.org/zap"
)

var ErrNoModels = errors.New("no models configured for router")

type RouterID = string

type LangRouter struct {
	routerID          RouterID
	Config            *LangRouterConfig
	chatModels        []*extmodel.LanguageModel
	chatStreamModels  []*extmodel.LanguageModel
	chatRouting       routing.LangModelRouting
	chatStreamRouting routing.LangModelRouting
	retry             *retry.ExpRetry
	tel               *telemetry.Telemetry
	logger            *zap.Logger
}

func NewLangRouter(cfg *LangRouterConfig, tel *telemetry.Telemetry) (*LangRouter, error) {
	chatModels, chatStreamModels, err := cfg.BuildModels(tel)
	if err != nil {
		return nil, err
	}

	chatRouting, chatStreamRouting, err := cfg.BuildRouting(chatModels, chatStreamModels)
	if err != nil {
		return nil, err
	}

	router := &LangRouter{
		routerID:          cfg.ID,
		Config:            cfg,
		chatModels:        chatModels,
		chatStreamModels:  chatStreamModels,
		retry:             cfg.BuildRetry(),
		chatRouting:       chatRouting,
		chatStreamRouting: chatStreamRouting,
		tel:               tel,
		logger:            tel.L().With(zap.String("routerID", cfg.ID)),
	}

	return router, err
}

func (r *LangRouter) ID() RouterID {
	return r.routerID
}

func (r *LangRouter) Chat(ctx context.Context, req *schemas.ChatRequest) (*schemas.ChatResponse, error) {
	if len(r.chatModels) == 0 {
		return nil, ErrNoModels
	}

	retryIterator := r.retry.Iterator()

	for retryIterator.HasNext() {
		modelIterator := r.chatRouting.Iterator()

		for {
			model, err := modelIterator.Next()

			if errors.Is(err, routing.ErrNoHealthyModels) {
				// no healthy model in the pool. Let's retry after some time
				break
			}

			langModel := model.(extmodel.LangModel)

			chatParams := req.Params(langModel.ID(), langModel.ModelName())

			resp, err := langModel.Chat(ctx, chatParams)
			if err != nil {
				r.logger.Warn(
					"Lang model failed processing chat request",
					zap.String("modelID", langModel.ID()),
					zap.String("provider", langModel.Provider()),
					zap.Error(err),
				)

				continue
			}

			resp.RouterID = r.routerID

			return resp, nil
		}

		// no providers were available to handle the request,
		//  so we have to wait a bit with a hope there is some available next time
		r.logger.Warn("No healthy model found to serve chat request, wait and retry")

		err := retryIterator.WaitNext(ctx)
		if err != nil {
			// something has cancelled the context
			return nil, err
		}
	}

	// if we reach this part, then we are in trouble
	r.logger.Error("No model was available to handle chat request")

	return nil, &schemas.ErrNoModelAvailable
}

func (r *LangRouter) ChatStream(
	ctx context.Context,
	req *schemas.ChatStreamRequest,
	respC chan<- *schemas.ChatStreamMessage,
) {
	if len(r.chatStreamModels) == 0 {
		respC <- schemas.NewChatStreamError(
			req.ID,
			r.routerID,
			schemas.NoModelConfigured,
			ErrNoModels.Error(),
			req.Metadata,
			&schemas.ReasonError,
		)

		return
	}

	retryIterator := r.retry.Iterator()

	for retryIterator.HasNext() {
		modelIterator := r.chatStreamRouting.Iterator()

	NextModel:
		for {
			model, err := modelIterator.Next()

			if errors.Is(err, routing.ErrNoHealthyModels) {
				// no healthy model in the pool. Let's retry after some time
				break
			}

			langModel := model.(extmodel.LangModel)
			chatParams := req.Params(langModel.ID(), langModel.ModelName())

			modelRespC, err := langModel.ChatStream(ctx, chatParams)
			if err != nil {
				r.logger.Error(
					"Lang model failed to create streaming chat request",
					zap.String("modelID", langModel.ID()),
					zap.String("provider", langModel.Provider()),
					zap.Error(err),
				)

				continue
			}

			for chunkResult := range modelRespC {
				err = chunkResult.Error()
				if err != nil {
					r.logger.Warn(
						"Lang model failed processing streaming chat request",
						zap.String("modelID", langModel.ID()),
						zap.String("provider", langModel.Provider()),
						zap.Error(err),
					)

					// It's challenging to hide an error in case of streaming chat as consumer apps
					//  may have already used all chunks we streamed this far (e.g. showed them to their users like OpenAI UI does),
					//  so we cannot easily restart that process from scratch
					respC <- schemas.NewChatStreamError(
						req.ID,
						r.routerID,
						schemas.ModelUnavailable,
						err.Error(),
						req.Metadata,
						nil,
					)

					continue NextModel
				}

				chunk := chunkResult.Chunk()

				respC <- schemas.NewChatStreamChunk(
					req.ID,
					r.routerID,
					req.Metadata,
					chunk,
				)
			}

			return
		}

		// no providers were available to handle the request,
		//  so we have to wait a bit with a hope there is some available next time
		r.logger.Warn("No healthy model found to serve streaming chat request, wait and retry")

		err := retryIterator.WaitNext(ctx)
		if err != nil {
			// something has cancelled the context
			respC <- schemas.NewChatStreamError(
				req.ID,
				r.routerID,
				schemas.UnknownError,
				err.Error(),
				req.Metadata,
				nil,
			)

			return
		}
	}

	// if we reach this part, then we are in trouble
	r.logger.Error(
		"No model was available to handle streaming chat request. " +
			"Try to configure more fallback models to avoid this",
	)

	respC <- schemas.NewChatStreamError(
		req.ID,
		r.routerID,
		schemas.ErrNoModelAvailable.Name,
		schemas.ErrNoModelAvailable.Message,
		req.Metadata,
		&schemas.ReasonError,
	)
}
