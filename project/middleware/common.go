package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"project/core"
	"strings"
	"time"
)

// Contoh implementasi logging yang lebih aman
func Logging[R any, S any](actionHandler core.ActionHandler[R, S], indentation int) core.ActionHandler[R, S] {
	return func(ctx context.Context, request R) (*S, error) {
		// Coba marshaling JSON, tangkap error jika gagal
		reqStr := fmt.Sprintf("%T", request)
		bytes, err := json.Marshal(request)
		if err == nil {
			reqStr = string(bytes)
		}

		printWithIndentation(fmt.Sprintf(">>> REQUEST          %s\n", reqStr), indentation)

		response, err := actionHandler(ctx, request)
		if err != nil {
			printWithIndentation(fmt.Sprintf(">>> RESPONSE ERROR  %s\n\n", err.Error()), indentation)
			printLine(indentation)
			return nil, err
		}

		// Coba marshaling JSON response, tangkap error jika gagal
		respStr := fmt.Sprintf("%T", response)
		if response != nil {
			bytes, err := json.Marshal(response)
			if err == nil {
				respStr = string(bytes)
			}
		}

		printWithIndentation(fmt.Sprintf(">>> RESPONSE SUCCESS %s\n\n", respStr), indentation)
		printLine(indentation)

		return response, nil
	}
}

func printLine(indentation int) {
	if indentation == 0 {
		fmt.Printf("----------------------------------------------------------------------------------------------------------\n")
	}
}

func printWithIndentation(message string, indentation int) {
	indentStr := strings.Repeat(" ", indentation)
	fmt.Printf("%s%s", indentStr, message)
}

func Retry[R any, S any](actionHandler core.ActionHandler[R, S], attempt int) core.ActionHandler[R, S] {
	return func(ctx context.Context, request R) (*S, error) {

		count := 1

		for {
			response, err := actionHandler(ctx, request)
			if err != nil {

				count++
				if count <= attempt {
					continue
				} else {
					return nil, err
				}

			}

			return response, nil
		}

	}
}

func Timing[R any, S any](actionHandler core.ActionHandler[R, S], label string) core.ActionHandler[R, S] {
	return func(ctx context.Context, request R) (*S, error) {
		start := time.Now()

		response, err := actionHandler(ctx, request)

		duration := time.Since(start)
		fmt.Printf("Request %s took %v\n", label, duration)

		return response, err
	}
}
