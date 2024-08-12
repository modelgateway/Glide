package http

import (
	"context"
	"sync"

	"github.com/EinStack/glide/pkg/router"

	"github.com/EinStack/glide/pkg/telemetry"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler = func(c *fiber.Ctx) error

// Swagger 101:
// - https://github.com/swaggo/swag/tree/master/example/celler

// LangChatHandler
//
//	@id				glide-language-chat
//	@Summary		Language Chat
//	@Description	Talk to different LLM Chat APIs via unified endpoint
//	@tags			Language
//	@Param			router	path	string						true	"Router ID"
//	@Param			payload	body	schemas.ChatRequest	true	"Request Data"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	schemas.ChatResponse
//	@Failure		400	{object}	schemas.Error
//	@Failure		404	{object}	schemas.Error
//	@Router			/v1/language/{router}/chat [POST]
func LangChatHandler(routerManager *router.Manager) Handler {
	return func(c *fiber.Ctx) error {
		if !c.Is("json") {
			return c.Status(fiber.StatusBadRequest).JSON(schema.ErrUnsupportedMediaType)
		}

		// Unmarshal request body
		req := schema.GetChatRequest()
		defer schema.ReleaseChatRequest(req)

		err := c.BodyParser(&req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(schema.NewPayloadParseErr(err))
		}

		// Get router ID from path
		routerID := c.Params("router")

		r, err := routerManager.GetLangRouter(routerID)
		if err != nil {
			httpErr := schema.FromErr(err)

			return c.Status(httpErr.Status).JSON(httpErr)
		}

		// Chat with router
		resp := schema.GetChatResponse()
		defer schema.ReleaseChatResponse(resp)

		resp, err = r.Chat(c.Context(), req)
		if err != nil {
			httpErr := schema.FromErr(err)

			return c.Status(httpErr.Status).JSON(httpErr)
		}

		// Return chat response
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func LangStreamRouterValidator(routerManager *router.Manager) Handler {
	return func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			routerID := c.Params("router")

			_, err := routerManager.GetLangRouter(routerID)
			if err != nil {
				httpErr := schema.FromErr(err)

				return c.Status(httpErr.Status).JSON(httpErr)
			}

			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	}
}

// LangStreamChatHandler
//
//	@id				glide-language-chat-stream
//	@Summary		Language Chat
//	@Description	Talk to different LLM Stream Chat APIs via a unified websocket endpoint
//	@tags			Language
//	@Param			router	    			path		string	true	"Router ID"
//	@Param			Connection				header		string	true	"Websocket Connection Type"
//	@Param			Upgrade 				header		string	true	"Upgrade header"
//	@Param			Sec-WebSocket-Key  		header		string	true	"Websocket Security Token"
//	@Param			Sec-WebSocket-Version  	header		string	true	"Websocket Security Token"
//	@Accept			json
//	@Success		101
//	@Failure		426
//	@Failure		404	{object}	schemas.Error
//	@Router			/v1/language/{router}/chatStream [GET]
func LangStreamChatHandler(tel *telemetry.Telemetry, routerManager *router.Manager) Handler {
	// TODO: expose websocket connection configs https://github.com/gofiber/contrib/tree/main/websocket
	return websocket.New(func(c *websocket.Conn) {
		routerID := c.Params("router")
		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index

		var (
			err error
			wg  sync.WaitGroup
		)

		chatStreamC := make(chan *schema.ChatStreamMessage)

		r, _ := routerManager.GetLangRouter(routerID)

		defer close(chatStreamC)
		defer c.Conn.Close()

		wg.Add(1)

		go func() {
			defer wg.Done()

			for chatStreamMsg := range chatStreamC {
				if err = c.WriteJSON(chatStreamMsg); err != nil {
					break
				}
			}
		}()

		for {
			var chatRequest schema.ChatStreamRequest

			if err = c.ReadJSON(&chatRequest); err != nil {
				// TODO: handle bad request schemas gracefully and return back validation errors
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					tel.L().Warn("Streaming Chat connection is closed", zap.Error(err), zap.String("routerID", routerID))
				}

				tel.L().Debug("Streaming chat connection is closed by client", zap.Error(err), zap.String("routerID", routerID))

				break
			}

			// TODO: handle termination gracefully
			wg.Add(1)

			go func(chatRequest schema.ChatStreamRequest) {
				defer wg.Done()

				r.ChatStream(context.Background(), &chatRequest, chatStreamC)
			}(chatRequest)
		}

		wg.Wait()
	})
}

// LangRoutersHandler
//
//	@id				glide-language-routers
//	@Summary		Language Router List
//	@Description	Retrieve list of configured active language routers and their configurations
//	@tags			Language
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	schemas.RouterListSchema
//	@Router			/v1/language/ [GET]
func LangRoutersHandler(routerManager *router.Manager) Handler {
	return func(c *fiber.Ctx) error {
		configuredRouters := routerManager.GetLangRouters()
		cfgs := make([]interface{}, 0, len(configuredRouters)) // opaque by design

		for _, r := range configuredRouters {
			cfgs = append(cfgs, r.Config)
		}

		return c.Status(fiber.StatusOK).JSON(schema.RouterListSchema{Routers: cfgs})
	}
}

// HealthHandler
//
//	@id			glide-health
//	@Summary	Gateway Health
//	@Description
//	@tags		Operations
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	schemas.HealthSchema
//	@Router		/v1/health/ [get]
func HealthHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(schema.HealthSchema{Healthy: true})
}

func NotFoundHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(schema.ErrRouteNotFound)
}
