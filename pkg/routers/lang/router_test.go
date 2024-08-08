package lang

import (
	"context"
	"testing"
	"time"

	ptesting "github.com/EinStack/glide/pkg/providers/testing"

	"github.com/EinStack/glide/pkg/models"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/resiliency/health"
	"github.com/EinStack/glide/pkg/resiliency/retry"

	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/routers/latency"
	"github.com/EinStack/glide/pkg/routers/routing"
	"github.com/EinStack/glide/pkg/telemetry"
	"github.com/stretchr/testify/require"
)

func TestLangRouter_Chat_PickFistHealthy(t *testing.T) {
	budget := health.NewErrorBudget(3, health.SEC)
	latConfig := latency.DefaultConfig()

	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Msg: "1"}, {Msg: "2"}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Msg: "1"}}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:         "test_router",
		retry:            retry.NewExpRetry(3, 2, 1*time.Second, nil),
		chatRouting:      routing.NewPriority(modelPool),
		chatModels:       langModels,
		chatStreamModels: langModels,
		tel:              telemetry.NewTelemetryMock(),
	}

	ctx := context.Background()
	req := schemas.NewChatFromStr("tell me a dad joke")

	for i := 0; i < 2; i++ {
		resp, err := router.Chat(ctx, req)

		require.Equal(t, "first", resp.ModelID)
		require.Equal(t, "test_router", resp.RouterID)
		require.NoError(t, err)
	}
}

func TestLangRouter_Chat_PickThirdHealthy(t *testing.T) {
	budget := health.NewErrorBudget(1, health.SEC)
	latConfig := latency.DefaultConfig()
	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Msg: "3"}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Msg: "4"}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"third",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Msg: "1"}, {Msg: "2"}}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	expectedModels := []string{"third", "third"}

	router := Router{
		routerID:          "test_router",
		retry:             retry.NewExpRetry(3, 2, 1*time.Second, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatStreamRouting: routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamModels:  langModels,
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	ctx := context.Background()
	req := schemas.NewChatFromStr("tell me a dad joke")

	for _, modelID := range expectedModels {
		resp, err := router.Chat(ctx, req)

		require.NoError(t, err)
		require.Equal(t, modelID, resp.ModelID)
		require.Equal(t, "test_router", resp.RouterID)
	}
}

func TestLangRouter_Chat_SuccessOnRetry(t *testing.T) {
	budget := health.NewErrorBudget(1, health.MILLI)
	latConfig := latency.DefaultConfig()
	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Msg: "2"}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Msg: "1"}}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_router",
		retry:             retry.NewExpRetry(3, 2, 1*time.Millisecond, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatStreamRouting: routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamModels:  langModels,
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	resp, err := router.Chat(context.Background(), schemas.NewChatFromStr("tell me a dad joke"))

	require.NoError(t, err)
	require.Equal(t, "first", resp.ModelID)
	require.Equal(t, "test_router", resp.RouterID)
}

func TestLangRouter_Chat_UnhealthyModelInThePool(t *testing.T) {
	budget := health.NewErrorBudget(1, health.MIN)
	latConfig := latency.DefaultConfig()
	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: clients.ErrProviderUnavailable}, {Msg: "3"}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Msg: "1"}, {Msg: "2"}}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_router",
		retry:             retry.NewExpRetry(3, 2, 1*time.Millisecond, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamModels:  langModels,
		chatStreamRouting: routing.NewPriority(modelPool),
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	for i := 0; i < 2; i++ {
		resp, err := router.Chat(context.Background(), schemas.NewChatFromStr("tell me a dad joke"))

		require.NoError(t, err)
		require.Equal(t, "second", resp.ModelID)
		require.Equal(t, "test_router", resp.RouterID)
	}
}

func TestLangRouter_Chat_AllModelsUnavailable(t *testing.T) {
	budget := health.NewErrorBudget(1, health.SEC)
	latConfig := latency.DefaultConfig()
	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Err: &schemas.ErrNoModelAvailable}}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewProviderMock(nil, []ptesting.RespMock{{Err: &schemas.ErrNoModelAvailable}, {Err: &schemas.ErrNoModelAvailable}}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_router",
		retry:             retry.NewExpRetry(1, 2, 1*time.Millisecond, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamModels:  langModels,
		chatStreamRouting: routing.NewPriority(modelPool),
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	_, err := router.Chat(context.Background(), schemas.NewChatFromStr("tell me a dad joke"))

	require.Error(t, err)
}

func TestLangRouter_ChatStream(t *testing.T) {
	budget := health.NewErrorBudget(3, health.SEC)
	latConfig := latency.DefaultConfig()

	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewStreamProviderMock(nil, []ptesting.RespStreamMock{
				ptesting.NewRespStreamMock(&[]ptesting.RespMock{
					{Msg: "Bill"},
					{Msg: "Gates"},
					{Msg: "entered"},
					{Msg: "the"},
					{Msg: "bar"},
				}),
			}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewStreamProviderMock(nil, []ptesting.RespStreamMock{
				ptesting.NewRespStreamMock(&[]ptesting.RespMock{
					{Msg: "Knock"},
					{Msg: "Knock"},
					{Msg: "joke"},
				}),
			}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_stream_router",
		retry:             retry.NewExpRetry(3, 2, 1*time.Second, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamRouting: routing.NewPriority(modelPool),
		chatStreamModels:  langModels,
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	ctx := context.Background()
	req := schemas.NewChatStreamFromStr("tell me a dad joke")
	respC := make(chan *schemas.ChatStreamMessage)

	defer close(respC)

	go router.ChatStream(ctx, req, respC)

	chunks := make([]string, 0, 5)

	for range 5 {
		select { //nolint:gosimple
		case message := <-respC:
			require.Nil(t, message.Error)
			require.NotNil(t, message.Chunk)
			require.NotNil(t, message.Chunk.ModelResponse.Message.Content)

			chunks = append(chunks, message.Chunk.ModelResponse.Message.Content)
		}
	}

	require.Equal(t, []string{"Bill", "Gates", "entered", "the", "bar"}, chunks)
}

func TestLangRouter_ChatStream_FailOnFirst(t *testing.T) {
	budget := health.NewErrorBudget(3, health.SEC)
	latConfig := latency.DefaultConfig()

	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewStreamProviderMock(nil, nil),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewStreamProviderMock(nil, []ptesting.RespStreamMock{
				ptesting.NewRespStreamMock(
					&[]ptesting.RespMock{
						{Msg: "Knock"},
						{Msg: "knock"},
						{Msg: "joke"},
					},
				),
			}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_stream_router",
		retry:             retry.NewExpRetry(3, 2, 1*time.Second, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamRouting: routing.NewPriority(modelPool),
		chatStreamModels:  langModels,
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	ctx := context.Background()
	req := schemas.NewChatStreamFromStr("tell me a dad joke")
	respC := make(chan *schemas.ChatStreamMessage)

	defer close(respC)

	go router.ChatStream(ctx, req, respC)

	chunks := make([]string, 0, 3)

	for range 3 {
		select { //nolint:gosimple
		case message := <-respC:
			require.Nil(t, message.Error)
			require.NotNil(t, message.Chunk.ModelResponse.Message.Content)
			require.NotNil(t, message.Chunk.ModelResponse.Message.Content)

			chunks = append(chunks, message.Chunk.ModelResponse.Message.Content)
		}
	}

	require.Equal(t, []string{"Knock", "knock", "joke"}, chunks)
}

func TestLangRouter_ChatStream_AllModelsUnavailable(t *testing.T) {
	budget := health.NewErrorBudget(1, health.SEC)
	latConfig := latency.DefaultConfig()

	langModels := []*models.LanguageModel{
		models.NewLangModel(
			"first",
			ptesting.NewStreamProviderMock(nil, []ptesting.RespStreamMock{
				ptesting.NewRespStreamMock(&[]ptesting.RespMock{
					{Err: clients.ErrProviderUnavailable},
				}),
			}),
			budget,
			*latConfig,
			1,
		),
		models.NewLangModel(
			"second",
			ptesting.NewStreamProviderMock(nil, []ptesting.RespStreamMock{
				ptesting.NewRespStreamMock(&[]ptesting.RespMock{
					{Err: clients.ErrProviderUnavailable},
				}),
			}),
			budget,
			*latConfig,
			1,
		),
	}

	modelPool := make([]models.Model, 0, len(langModels))
	for _, model := range langModels {
		modelPool = append(modelPool, model)
	}

	router := Router{
		routerID:          "test_router",
		retry:             retry.NewExpRetry(1, 2, 1*time.Millisecond, nil),
		chatRouting:       routing.NewPriority(modelPool),
		chatModels:        langModels,
		chatStreamModels:  langModels,
		chatStreamRouting: routing.NewPriority(modelPool),
		tel:               telemetry.NewTelemetryMock(),
		logger:            telemetry.NewLoggerMock(),
	}

	respC := make(chan *schemas.ChatStreamMessage)
	defer close(respC)

	go router.ChatStream(context.Background(), schemas.NewChatStreamFromStr("tell me a dad joke"), respC)

	errs := make([]string, 0, 3)

	for range 3 {
		result := <-respC
		require.Nil(t, result.Chunk)
		require.NotNil(t, result.Error)

		errs = append(errs, result.Error.Name)
	}

	require.Equal(t, []string{schemas.ModelUnavailable, schemas.ModelUnavailable, schemas.AllModelsUnavailable}, errs)
}
