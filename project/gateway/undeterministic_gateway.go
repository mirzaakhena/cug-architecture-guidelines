package gateway

import (
	"context"
	"project/core"
	"time"

	"github.com/google/uuid"
)

// GenerateUUIDReq is the request for generating a UUID
type GenerateUUIDReq struct{}

// GenerateUUIDRes is the response for generating a UUID
type GenerateUUIDRes struct {
	UUID string
}

// GenerateUUID is the gateway for generating a UUID
type GenerateUUID = core.ActionHandler[GenerateUUIDReq, GenerateUUIDRes]

// ImplGenerateUUID implements the GenerateUUID gateway
func ImplGenerateUUID() GenerateUUID {
	return func(ctx context.Context, req GenerateUUIDReq) (*GenerateUUIDRes, error) {
		return &GenerateUUIDRes{
			UUID: uuid.New().String(),
		}, nil
	}
}

// GetCurrentTimeReq is the request for getting the current time
type GetCurrentTimeReq struct{}

// GetCurrentTimeRes is the response for getting the current time
type GetCurrentTimeRes struct {
	Now time.Time
}

// GetCurrentTime is the gateway for getting the current time
type GetCurrentTime = core.ActionHandler[GetCurrentTimeReq, GetCurrentTimeRes]

// ImplGetCurrentTime implements the GetCurrentTime gateway
func ImplGetCurrentTime() GetCurrentTime {
	return func(ctx context.Context, req GetCurrentTimeReq) (*GetCurrentTimeRes, error) {
		return &GetCurrentTimeRes{
			Now: time.Now(),
		}, nil
	}
}
