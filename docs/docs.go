// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Glide Community",
            "url": "https://github.com/modelgateway/glide"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "https://github.com/modelgateway/glide/blob/develop/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v1/health/": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Operations"
                ],
                "summary": "Gateway Health",
                "operationId": "glide-health",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/http.HealthSchema"
                        }
                    }
                }
            }
        },
        "/v1/language/": {
            "get": {
                "description": "Retrieve list of configured language routers and their configurations",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Language"
                ],
                "summary": "Language Router List",
                "operationId": "glide-language-routers",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/http.RouterListSchema"
                        }
                    }
                }
            }
        },
        "/v1/language/{router}/chat": {
            "post": {
                "description": "Talk to different LLMs Chat API via unified endpoint",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Language"
                ],
                "summary": "Language Chat",
                "operationId": "glide-language-chat",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Router ID",
                        "name": "router",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Request Data",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/schemas.UnifiedChatRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/schemas.UnifiedChatResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/http.ErrorSchema"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/http.ErrorSchema"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "clients.ClientConfig": {
            "type": "object",
            "properties": {
                "timeout": {
                    "type": "integer"
                }
            }
        },
        "http.ErrorSchema": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "http.HealthSchema": {
            "type": "object",
            "properties": {
                "healthy": {
                    "type": "boolean"
                }
            }
        },
        "http.RouterListSchema": {
            "type": "object",
            "properties": {
                "routers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/routers.LangRouterConfig"
                    }
                }
            }
        },
        "openai.Config": {
            "type": "object",
            "required": [
                "baseUrl",
                "chatEndpoint",
                "model"
            ],
            "properties": {
                "baseUrl": {
                    "type": "string"
                },
                "chatEndpoint": {
                    "type": "string"
                },
                "defaultParams": {
                    "$ref": "#/definitions/openai.Params"
                },
                "model": {
                    "type": "string"
                }
            }
        },
        "openai.Params": {
            "type": "object",
            "properties": {
                "frequency_penalty": {
                    "type": "integer"
                },
                "logit_bias": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "number"
                    }
                },
                "max_tokens": {
                    "type": "integer"
                },
                "n": {
                    "type": "integer"
                },
                "presence_penalty": {
                    "type": "integer"
                },
                "response_format": {
                    "description": "TODO: should this be a part of the chat request API?"
                },
                "seed": {
                    "type": "integer"
                },
                "stop": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "temperature": {
                    "type": "number"
                },
                "tool_choice": {},
                "tools": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "top_p": {
                    "type": "number"
                },
                "user": {
                    "type": "string"
                }
            }
        },
        "providers.LangModelConfig": {
            "type": "object",
            "required": [
                "id"
            ],
            "properties": {
                "client": {
                    "$ref": "#/definitions/clients.ClientConfig"
                },
                "enabled": {
                    "description": "Is the model enabled?",
                    "type": "boolean"
                },
                "error_budget": {
                    "type": "string"
                },
                "id": {
                    "description": "Model instance ID (unique in scope of the router)",
                    "type": "string"
                },
                "openai": {
                    "$ref": "#/definitions/openai.Config"
                }
            }
        },
        "retry.ExpRetryConfig": {
            "type": "object",
            "properties": {
                "base_multiplier": {
                    "type": "integer"
                },
                "max_delay": {
                    "type": "integer"
                },
                "max_retries": {
                    "type": "integer"
                },
                "min_delay": {
                    "type": "integer"
                }
            }
        },
        "routers.LangRouterConfig": {
            "type": "object",
            "required": [
                "models",
                "routers"
            ],
            "properties": {
                "enabled": {
                    "description": "Is router enabled?",
                    "type": "boolean"
                },
                "models": {
                    "description": "the list of models that could handle requests",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/providers.LangModelConfig"
                    }
                },
                "retry": {
                    "description": "retry when no healthy model is available to router",
                    "allOf": [
                        {
                            "$ref": "#/definitions/retry.ExpRetryConfig"
                        }
                    ]
                },
                "routers": {
                    "description": "Unique router ID",
                    "type": "string"
                },
                "strategy": {
                    "description": "strategy on picking the next model to serve the request",
                    "allOf": [
                        {
                            "$ref": "#/definitions/routing.Strategy"
                        }
                    ]
                }
            }
        },
        "routing.Strategy": {
            "type": "string",
            "enum": [
                "priority",
                "round-robin",
                "least_latency"
            ],
            "x-enum-varnames": [
                "Priority",
                "RoundRobin",
                "LeastLatency"
            ]
        },
        "schemas.ChatMessage": {
            "type": "object",
            "properties": {
                "content": {
                    "description": "The content of the message.",
                    "type": "string"
                },
                "name": {
                    "description": "The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,\nwith a maximum length of 64 characters.",
                    "type": "string"
                },
                "role": {
                    "description": "The role of the author of this message. One of system, user, or assistant.",
                    "type": "string"
                }
            }
        },
        "schemas.ProviderResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "$ref": "#/definitions/schemas.ChatMessage"
                },
                "responseId": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "tokenCount": {
                    "$ref": "#/definitions/schemas.TokenCount"
                }
            }
        },
        "schemas.TokenCount": {
            "type": "object",
            "properties": {
                "promptTokens": {
                    "type": "number"
                },
                "responseTokens": {
                    "type": "number"
                },
                "totalTokens": {
                    "type": "number"
                }
            }
        },
        "schemas.UnifiedChatRequest": {
            "type": "object",
            "properties": {
                "message": {
                    "$ref": "#/definitions/schemas.ChatMessage"
                },
                "messageHistory": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/schemas.ChatMessage"
                    }
                }
            }
        },
        "schemas.UnifiedChatResponse": {
            "type": "object",
            "properties": {
                "cached": {
                    "type": "boolean"
                },
                "created": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "model": {
                    "type": "string"
                },
                "modelResponse": {
                    "$ref": "#/definitions/schemas.ProviderResponse"
                },
                "model_id": {
                    "type": "string"
                },
                "provider": {
                    "type": "string"
                },
                "router": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:9099",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Glide Gateway",
	Description:      "API documentation for Glide, an open-source lightweight high-performance model gateway",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
