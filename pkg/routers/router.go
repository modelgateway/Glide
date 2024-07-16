package routers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/EinStack/glide/pkg/cache"
	"github.com/EinStack/glide/pkg/routers/retry"
	"go.uber.org/zap"

	"github.com/EinStack/glide/pkg/providers"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/routers/routing"

	"github.com/EinStack/glide/pkg/api/schemas"
)

var ErrNoModels = errors.New("no models configured for router")

type RouterID = string

type LangRouter struct {
	routerID          RouterID
	Config            *LangRouterConfig
	chatModels        []*providers.LanguageModel
	chatStreamModels  []*providers.LanguageModel
	chatRouting       routing.LangModelRouting
	chatStreamRouting routing.LangModelRouting
	retry             *retry.ExpRetry
	tel               *telemetry.Telemetry
	logger            *zap.Logger
	cache             *cache.MemoryCache
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
		cache:             cache.NewMemoryCache(),
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

	// Generate cache key
	cacheKey := req.Message.Content
	if cachedResponse, found := r.cache.Get(cacheKey); found {
		log.Println("found cached response and returning: ", cachedResponse)
		if response, ok := cachedResponse.(*schemas.ChatResponse); ok {
			return response, nil
		} else {
			log.Println("Failed to cast cached response to ChatResponse")
		}
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

			langModel := model.(providers.LangModel)

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

			// Store response in cache
			r.cache.Set(cacheKey, resp)

			return resp, nil
		}

		r.logger.Warn("No healthy model found to serve chat request, wait and retry")

		err := retryIterator.WaitNext(ctx)
		if err != nil {
			// something has cancelled the context
			return nil, err
		}
	}

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

	cacheKey := req.Message.Content
	if streamingCacheEntry, found := r.cache.Get(cacheKey); found {
		if entry, ok := streamingCacheEntry.(*schemas.StreamingCacheEntry); ok {
			for _, chunkKey := range entry.ResponseChunks {
				if cachedChunk, found := r.cache.Get(chunkKey); found {
					if chunk, ok := cachedChunk.(*schemas.ChatStreamChunk); ok {
						respC <- schemas.NewChatStreamChunk(
							req.ID,
							r.routerID,
							req.Metadata,
							chunk,
						)
					} else {
						log.Println("Failed to cast cached chunk to ChatStreamChunk")
					}
				}
			}

			if entry.Complete {
				return
			}
		} else {
			log.Println("Failed to cast cached entry to StreamingCacheEntry")
		}
	} else {
		streamingCacheEntry := &schemas.StreamingCacheEntry{
			Key:            cacheKey,
			Query:          req.Message.Content,
			ResponseChunks: []string{},
			Complete:       false,
		}
		r.cache.Set(cacheKey, streamingCacheEntry)
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

			langModel := model.(providers.LangModel)
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

			buffer := []schemas.ChatStreamChunk{}
			for chunkResult := range modelRespC {
				err = chunkResult.Error()
				if err != nil {
					r.logger.Warn(
						"Lang model failed processing streaming chat request",
						zap.String("modelID", langModel.ID()),
						zap.String("provider", langModel.Provider()),
						zap.Error(err),
					)

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
				buffer = append(buffer, *chunk)
				respC <- schemas.NewChatStreamChunk(
					req.ID,
					r.routerID,
					req.Metadata,
					chunk,
				)

				if len(buffer) >= 1048 { // Define bufferSize as per your requirement
					chunkKey := fmt.Sprintf("%s-chunk-%d", cacheKey, len(buffer))
					r.cache.Set(chunkKey, &schemas.StreamingCacheEntryChunk{
						Key:     chunkKey,
						Index:   len(buffer),
						Content: *chunk,
					})
					streamingCacheEntry := schemas.StreamingCacheEntry{}
					streamingCacheEntry.ResponseChunks = append(streamingCacheEntry.ResponseChunks, chunkKey)
					buffer = buffer[:0] // Reset buffer
					r.cache.Set(cacheKey, streamingCacheEntry)
				}
			}

			if len(buffer) > 0 {
				chunkKey := fmt.Sprintf("%s-chunk-%d", cacheKey, len(buffer))
				r.cache.Set(chunkKey, &schemas.StreamingCacheEntryChunk{
					Key:     chunkKey,
					Index:   len(buffer),
					Content: buffer[0], // Assuming buffer has at least one element
				})
				streamingCacheEntry := schemas.StreamingCacheEntry{}
				streamingCacheEntry.ResponseChunks = append(streamingCacheEntry.ResponseChunks, chunkKey)
				buffer = buffer[:0] // Reset buffer
				r.cache.Set(cacheKey, streamingCacheEntry)
			}

			streamingCacheEntry := schemas.StreamingCacheEntry{}
			streamingCacheEntry.Complete = true
			r.cache.Set(cacheKey, streamingCacheEntry)

			return
		}

		r.logger.Warn("No healthy model found to serve streaming chat request, wait and retry")

		err := retryIterator.WaitNext(ctx)
		if err != nil {
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
