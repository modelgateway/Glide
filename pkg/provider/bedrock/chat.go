package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EinStack/glide/pkg/api/schemas"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// ChatRequest is a Bedrock-specific request schema
type ChatRequest struct {
	Messages             string               `json:"inputText"`
	TextGenerationConfig TextGenerationConfig `json:"textGenerationConfig"`
}

func (r *ChatRequest) ApplyParams(params *schemas.ChatParams) {
	// message history not yet supported for AWS models
	// TODO: do something about lack of message history. Maybe just concatenate all messages?
	// 	in any case, this is not a way to go to ignore message history
	message := params.Messages[len(params.Messages)-1]

	r.Messages = fmt.Sprintf("Role: %s, Content: %s", message.Role, message.Content)
}

type TextGenerationConfig struct {
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"topP"`
	MaxTokenCount int      `json:"maxTokenCount"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		TextGenerationConfig: TextGenerationConfig{
			MaxTokenCount: cfg.DefaultParams.MaxTokens,
			StopSequences: cfg.DefaultParams.StopSequence,
			Temperature:   cfg.DefaultParams.Temperature,
			TopP:          cfg.DefaultParams.TopP,
		},
	}
}

// Chat sends a chat request to the specified bedrock model.
func (c *Client) Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error) {
	// Create a new chat request
	// TODO: consider using objectpool to optimize memory allocation
	chatReq := *c.chatRequestTemplate // hoping to get a copy of the template
	chatReq.ApplyParams(params)

	chatResponse, err := c.doChatRequest(ctx, &chatReq)
	if err != nil {
		return nil, err
	}

	return chatResponse, nil
}

func (c *Client) doChatRequest(ctx context.Context, payload *ChatRequest) (*schemas.ChatResponse, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal chat request payload: %w", err)
	}

	result, err := c.bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.config.ModelName),
		ContentType: aws.String("application/json"),
		Body:        rawPayload,
	})
	if err != nil {
		c.telemetry.Logger.Error("Error: Couldn't invoke model. Here's why: %v\n", zap.Error(err))
		return nil, err
	}

	var bedrockCompletion ChatCompletion

	err = json.Unmarshal(result.Body, &bedrockCompletion)
	if err != nil {
		c.telemetry.Logger.Error("failed to parse bedrock chat response", zap.Error(err))

		return nil, err
	}

	modelResult := bedrockCompletion.Results[0]

	if len(modelResult.OutputText) == 0 {
		return nil, ErrEmptyResponse
	}

	response := schemas.ChatResponse{
		ID:        uuid.NewString(),
		Created:   int(time.Now().Unix()),
		Provider:  providerName,
		ModelName: c.config.ModelName,
		Cached:    false,
		ModelResponse: schemas.ModelResponse{
			Message: schemas.ChatMessage{
				Role:    "assistant",
				Content: modelResult.OutputText,
			},
			TokenUsage: schemas.TokenUsage{
				// TODO: what would happen if there is a few responses? We need to sum that up
				PromptTokens:   modelResult.TokenCount,
				ResponseTokens: -1,
				TotalTokens:    modelResult.TokenCount,
			},
		},
	}

	return &response, nil
}
