package core

import "context"

type ContextKey string

type ActionHandler[REQUEST any, RESPONSE any] func(ctx context.Context, req REQUEST) (*RESPONSE, error)

type MiddlewareHandler[REQUEST any, RESPONSE any] func(actionHandler ActionHandler[REQUEST, RESPONSE]) ActionHandler[REQUEST, RESPONSE]

type InternalServerError struct {
	error
}

func NewInternalServerError(err error) error {
	return InternalServerError{
		error: err,
	}
}

func (a InternalServerError) Error() string {
	return string(a.error.Error())
}

type ErrorWithData struct {
	error
	Data any
}

func NewErrorWithData(err error, data any) error {
	return ErrorWithData{
		error: err,
		Data:  data,
	}
}

func GetDataFromContext[dataType any](ctx context.Context, key ContextKey, defaultValue ...dataType) dataType {
	var x dataType
	data, ok := ctx.Value(key).(dataType)
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return x
	}
	return data
}

func AttachDataToContext[dataType any](ctx context.Context, key ContextKey, data dataType) context.Context {
	return context.WithValue(ctx, key, data)
}
