package schema

import (
	"sync"
)

var chatRequestPool = sync.Pool{
	New: func() interface{} {
		return &ChatRequest{}
	},
}

var chatResponsePool = sync.Pool{
	New: func() interface{} {
		return &ChatResponse{}
	},
}

// GetChatRequest get objects from the pool
func GetChatRequest() *ChatRequest {
	return chatRequestPool.Get().(*ChatRequest)
}

// ReleaseChatRequest release objects from the pool
func ReleaseChatRequest(req *ChatRequest) {
	*req = ChatRequest{}
	chatRequestPool.Put(req)
}

// GetChatResponse get objects from the pool
func GetChatResponse() *ChatResponse {
	return chatResponsePool.Get().(*ChatResponse)
}

// ReleaseChatResponse release objects from the pool
func ReleaseChatResponse(res *ChatResponse) {
	*res = ChatResponse{}
	chatResponsePool.Put(res)
}
