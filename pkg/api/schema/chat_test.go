package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func ToSlice(messageHistory []ChatMessage) []string {
	history := make([]string, 0, len(messageHistory))

	for _, message := range messageHistory {
		history = append(history, message.Content)
	}

	return history
}

// TestChatRequest_DefaultParams tests param creation for a request
// that doesn't have any override for a given model ID/name
func TestChatRequest_DefaultParams(t *testing.T) {
	backstory := "You are talking to a guy who won an ACMP contest in 2015"
	defaultMessage := "When did I win an ACMP contest?"

	modelID := "my-openai-model"
	myModelMessage := "When did he win the contest?Be concise"

	secondModelID := "my-other-model"
	secondModelName := "command-r"

	chatReq := ChatRequest{
		Message: ChatMessage{
			Role:    "user",
			Content: defaultMessage,
		},
		MessageHistory: []ChatMessage{
			{
				Role:    "system",
				Content: backstory,
			},
		},
		OverrideParams: &map[string]ModelParamsOverride{
			modelID: {
				Message: ChatMessage{
					Role:    "user",
					Content: myModelMessage,
				},
			},
		},
	}

	params := chatReq.Params(secondModelID, secondModelName)

	require.Equal(t, []string{backstory, defaultMessage}, ToSlice(params.Messages))
}

// TestChatRequest_ModelIDOverride tests param creation for a request
// that has a param override for a modelID
func TestChatRequest_ModelIDOverride(t *testing.T) {
	backstory := "You are talking to a guy who won an ACMP contest in 2015"
	defaultMessage := "When did I win an ACMP contest?"

	modelID := "my-openai-model"
	modelName := "gpt-4"
	myModelMessage := "When did he win the contest?Be concise"

	chatReq := ChatRequest{
		Message: ChatMessage{
			Role:    "user",
			Content: defaultMessage,
		},
		MessageHistory: []ChatMessage{
			{
				Role:    "system",
				Content: backstory,
			},
		},
		OverrideParams: &map[string]ModelParamsOverride{
			modelID: {
				Message: ChatMessage{
					Role:    "user",
					Content: myModelMessage,
				},
			},
		},
	}

	params := chatReq.Params(modelID, modelName)

	require.Equal(t, []string{backstory, myModelMessage}, ToSlice(params.Messages))
}

// TestChatRequest_ModelNameOverride tests param creation for a request
// that has a param override for a modelName
func TestChatRequest_ModelNameOverride(t *testing.T) {
	backstory := "You are talking to a guy who won an ACMP contest in 2015"
	defaultMessage := "When did I win an ACMP contest?"

	modelID := "my-openai-model"
	modelName := "gpt-4"
	myModelMessage := "When did he win the contest?Be concise"

	chatReq := ChatRequest{
		Message: ChatMessage{
			Role:    "user",
			Content: defaultMessage,
		},
		MessageHistory: []ChatMessage{
			{
				Role:    "system",
				Content: backstory,
			},
		},
		OverrideParams: &map[string]ModelParamsOverride{
			modelName: {
				Message: ChatMessage{
					Role:    "user",
					Content: myModelMessage,
				},
			},
		},
	}

	params := chatReq.Params(modelID, modelName)

	require.Equal(t, []string{backstory, myModelMessage}, ToSlice(params.Messages))
}

// TestChatRequest_ModelNameOverride tests param creation for a request
// that has a param override for both modelName & modelID
func TestChatRequest_ModelNameIDOverride(t *testing.T) {
	backstory := "You are talking to a guy who won an ACMP contest in 2015"
	defaultMessage := "When did I win an ACMP contest?"

	modelID := "my-openai-model"
	modelName := "gpt-4"
	myModelIDMessage := "When did he win the contest?Be concise"
	myModelNameMessage := "When did he win the contest? Answer like Illya would"

	chatReq := ChatRequest{
		Message: ChatMessage{
			Role:    "user",
			Content: defaultMessage,
		},
		MessageHistory: []ChatMessage{
			{
				Role:    "system",
				Content: backstory,
			},
		},
		OverrideParams: &map[string]ModelParamsOverride{
			modelName: {
				Message: ChatMessage{
					Role:    "user",
					Content: myModelNameMessage,
				},
			},
			modelID: {
				Message: ChatMessage{
					Role:    "user",
					Content: myModelIDMessage,
				},
			},
		},
	}

	params := chatReq.Params(modelID, modelName)

	require.Equal(t, []string{backstory, myModelIDMessage}, ToSlice(params.Messages))
}
