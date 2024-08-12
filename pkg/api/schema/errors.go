package schema

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ErrorName = string

var (
	UnsupportedMediaType ErrorName = "unsupported_media_type"
	RouteNotFound        ErrorName = "route_not_found"
	PayloadParseError    ErrorName = "payload_parse_error"
	RouterNotFound       ErrorName = "router_not_found"
	NoModelConfigured    ErrorName = "no_model_configured"
	ModelUnavailable     ErrorName = "model_unavailable"
	AllModelsUnavailable ErrorName = "all_models_unavailable"
	UnknownError         ErrorName = "unknown_error"
)

// Error / Error contains more context than the built-in error type,
// so we know information like error code and message that are useful to propagate to clients
type Error struct {
	Status  int    `json:"-"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

var _ error = (*Error)(nil)

// Error returns the error message.
func (e *Error) Error() string {
	return fmt.Sprintf("Error (%s): %s", e.Name, e.Message)
}

func NewError(status int, name string, message string) Error {
	return Error{Status: status, Name: name, Message: message}
}

var ErrUnsupportedMediaType = NewError(
	fiber.StatusBadRequest,
	UnsupportedMediaType,
	"application/json is the only supported media type",
)

var ErrRouteNotFound = NewError(
	fiber.StatusNotFound,
	RouteNotFound,
	"requested route is not found or method is not allowed",
)

var ErrRouterNotFound = NewError(fiber.StatusNotFound, RouterNotFound, "router is not found")

var ErrNoModelAvailable = NewError(
	503,
	AllModelsUnavailable,
	"all providers are unavailable",
)

func NewPayloadParseErr(err error) Error {
	return NewError(
		fiber.StatusBadRequest,
		PayloadParseError,
		err.Error(),
	)
}

func FromErr(err error) Error {
	if apiErr, ok := err.(*Error); ok {
		return *apiErr
	}

	return NewError(
		fiber.StatusInternalServerError,
		UnknownError,
		err.Error(),
	)
}
