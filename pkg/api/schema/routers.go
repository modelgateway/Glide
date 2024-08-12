package schema

// RouterListSchema returns list of active configured routers.
//
//	Routers config is exposed as an opaque value to indicate that user services must not use it to base any logic on it.
//	The endpoint is used for debugging/informational reasons
type RouterListSchema struct {
	Routers []interface{} `json:"routers"`
}
