package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"project/core"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Status   string  `json:"status"`
	Error    *string `json:"error"`
	Data     any     `json:"data"`
	Metadata any     `json:"metadata,omitempty"`
}

func internalServerError(w http.ResponseWriter, err error) {
	msg := errors.New("internal server error").Error()
	log.Println(err.Error()) // TODO create separate log here to make alert
	WriteJSON(w, http.StatusInternalServerError, Response{
		Status: "failed",
		Error:  &msg,
		Data:   nil,
	})
}

func badRequestError(w http.ResponseWriter, err error) {
	msg := err.Error()
	WriteJSON(w, http.StatusBadRequest, Response{
		Status: "failed",
		Error:  &msg,
		Data:   nil,
	})
}

func Success(w http.ResponseWriter, data any) {
	response := Response{
		Status: "success",
		Error:  nil,
		Data:   data,
	}

	WriteJSON(w, http.StatusOK, response)
}

func Fail(w http.ResponseWriter, err error) {
	var internalError core.InternalServerError
	if errors.As(err, &internalError) {
		internalServerError(w, err)
		return
	}

	badRequestError(w, err)
}

func WriteJSON(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// customDecoder wraps the standard JSON decoder to ignore fields tagged with "-"
type customDecoder struct {
	*json.Decoder
}

// Token returns the next JSON token, skipping fields tagged with "-"
func (d *customDecoder) Token() (json.Token, error) {
	token, err := d.Decoder.Token()
	if err != nil {
		return nil, err
	}

	// If the token is a field name (string), check if it should be ignored
	if str, ok := token.(string); ok {
		// Get the type information of the target struct
		val := reflect.ValueOf(d.Decoder).Elem().FieldByName("d").Elem().FieldByName("errorContext").FieldByName("typ")
		if val.IsValid() && val.Kind() == reflect.Ptr {
			typ := val.Elem().Type()
			if typ.Kind() == reflect.Struct {
				// Look for the field and check its JSON tag
				if field, exists := typ.FieldByName(str); exists {
					tag := field.Tag.Get("json")
					if tag == "-" {
						// Skip this field and its value
						if _, err := d.Token(); err != nil {
							return nil, err
						}
						return d.Token()
					}
				}
			}
		}
	}
	return token, nil
}

func ParseJSON[PayloadType any](w http.ResponseWriter, r *http.Request) (PayloadType, bool) {
	var x PayloadType

	decoder := &customDecoder{json.NewDecoder(r.Body)}

	if err := decoder.Decode(&x); err != nil {
		// if err := json.NewDecoder(r.Body).Decode(&x); err != nil {
		badRequestError(w, fmt.Errorf("invalid request body %v", err.Error()))
		return x, false
	}
	return x, true
}

func GetQueryInt(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetQueryFloat(r *http.Request, key string, defaultValue float64) float64 {
	valueStr := r.URL.Query().Get(key)
	if valueStr != "" {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return value
		}
	}
	return defaultValue
}

func GetQueryString(r *http.Request, key string, defaultValue string) string {
	if value := r.URL.Query().Get(key); value != "" {
		return value
	}
	return defaultValue
}

func GetQueryBoolean(r *http.Request, key string, defaultValue bool) bool {
	valueStr := r.URL.Query().Get(key)
	if valueStr != "" {
		lowerValue := strings.ToLower(valueStr)
		if lowerValue == "true" || lowerValue == "1" || lowerValue == "yes" {
			return true
		}
		if lowerValue == "false" || lowerValue == "0" || lowerValue == "no" {
			return false
		}
	}
	return defaultValue
}

func HandleUsecase[A any, B any](ctx context.Context, w http.ResponseWriter, useCase core.ActionHandler[A, B], req A) {
	response, err := useCase(ctx, req)
	if err != nil {
		Fail(w, err)
		return
	}
	Success(w, response)
}

func ExtractRequest[RequestType any](w http.ResponseWriter, r *http.Request, url string, f ...func(key string) (any, error)) (RequestType, bool) {
	var data RequestType
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(&data).Elem()

	// Create a single custom function that handles all keys
	var customFunc func(key string) (any, error)
	if len(f) > 0 {
		customFunc = f[0]
	}

	// Handle body
	bodyField, bodyFound := findTaggedField(t, "http", "body")
	if bodyFound {
		if bodyField.Type.Kind() != reflect.Struct {
			Fail(w, fmt.Errorf("field with http:\"body\" tag must be a struct"))
			return data, false
		}

		bodyValue := reflect.New(bodyField.Type).Interface()
		if err := json.NewDecoder(r.Body).Decode(bodyValue); err != nil {
			Fail(w, fmt.Errorf("failed to parse request body: %v", err))
			return data, false
		}
		v.FieldByIndex(bodyField.Index).Set(reflect.ValueOf(bodyValue).Elem())
	}

	// Handle path, query, and context parameters
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("http")
		if tag == "" {
			continue
		}

		switch {
		case tag == "path":
			pathValue := r.PathValue(field.Tag.Get("json"))
			if err := setField(v.Field(i), pathValue); err != nil {
				Fail(w, fmt.Errorf("failed to set path value: %v", err))
				return data, false
			}
		case tag == "query":
			queryKey := field.Tag.Get("json")
			switch field.Type.Kind() {
			case reflect.Int:
				value := GetQueryInt(r, queryKey, 0)
				v.Field(i).SetInt(int64(value))
			case reflect.Float64:
				value := GetQueryFloat(r, queryKey, 0)
				v.Field(i).SetFloat(value)
			case reflect.String:
				value := GetQueryString(r, queryKey, "")
				v.Field(i).SetString(value)
			default:
				Fail(w, fmt.Errorf("unsupported type for query parameter: %v", field.Type.Kind()))
				return data, false
			}
		case tag == "context":
			contextKey := field.Tag.Get("json")
			contextValue := core.GetDataFromContext[any](r.Context(), core.ContextKey(contextKey))
			if err := setFieldFromContext(v.Field(i), contextValue); err != nil {
				Fail(w, fmt.Errorf("failed to set context value: %v", err))
				return data, false
			}

		case tag == "now":
			if field.Type != reflect.TypeOf(time.Time{}) {
				Fail(w, fmt.Errorf("field with http:\"now\" tag must be of type time.Time"))
				return data, false
			}
			v.Field(i).Set(reflect.ValueOf(time.Now()))

		case strings.HasPrefix(tag, "func("):
			funcKey := strings.TrimSuffix(strings.TrimPrefix(tag, "func("), ")")
			if customFunc != nil {
				result, err := customFunc(funcKey)
				if err != nil {
					Fail(w, fmt.Errorf("error calling function for key %s: %v", funcKey, err))
					return data, false
				}
				if err := setFieldFromAny(v.Field(i), result); err != nil {
					Fail(w, fmt.Errorf("failed to set func value for key %s: %v", funcKey, err))
					return data, false
				}
			} else {
				Fail(w, fmt.Errorf("no custom function provided for key %s", funcKey))
				return data, false
			}
		}

	}

	return data, true
}

func findTaggedField(t reflect.Type, key, value string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get(key); tag == value {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		field.SetInt(int64(intValue))
	// Add more cases for other types as needed
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

func setFieldFromContext(field reflect.Value, value any) error {
	if value == nil {
		return nil // Skip setting if the context value is nil
	}

	switch field.Kind() {
	case reflect.String:
		if strValue, ok := value.(string); ok {
			field.SetString(strValue)
		} else {
			return fmt.Errorf("context value is not a string")
		}
	case reflect.Int:
		if intValue, ok := value.(int); ok {
			field.SetInt(int64(intValue))
		} else {
			return fmt.Errorf("context value is not an int")
		}
	// Add more cases for other types as needed
	default:
		return fmt.Errorf("unsupported field type for context value: %v", field.Kind())
	}
	return nil
}

func setFieldFromAny(field reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	valueReflect := reflect.ValueOf(value)

	if field.Type() == valueReflect.Type() {
		field.Set(valueReflect)
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, ok := value.(int64)
		if !ok {
			if floatValue, ok := value.(float64); ok {
				intValue = int64(floatValue)
			} else {
				return fmt.Errorf("cannot convert %v to int64", value)
			}
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, ok := value.(uint64)
		if !ok {
			if floatValue, ok := value.(float64); ok {
				uintValue = uint64(floatValue)
			} else {
				return fmt.Errorf("cannot convert %v to uint64", value)
			}
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, ok := value.(float64)
		if !ok {
			return fmt.Errorf("cannot convert %v to float64", value)
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("cannot convert %v to bool", value)
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}
