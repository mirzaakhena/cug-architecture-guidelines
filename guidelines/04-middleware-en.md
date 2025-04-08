## Rules for MIDDLEWARE Implementation

### ✅ REQUIRED
- Implement middleware as higher-order functions
- Accept and return `core.ActionHandler[R, S]`
- Execute the usecase with the same parameters as the middleware
- Maintain the usecase's functional signature
- Use generic type parameters for flexibility

### ❌ FORBIDDEN
- NEVER create middleware that depends on specific usecase implementations
- NEVER create permanent side effects from middleware
- NEVER fundamentally change the usecase's behavior

### 💡 IMPORTANT TO REMEMBER
- Middleware is optional and conditional—not all usecases need all middleware
- The order of middleware application matters and affects behavior
- Middleware is ideal for cross-cutting concerns like logging, monitoring, authentication
- Common middleware patterns: logging, timing, transaction, retry, caching, authorization
- Consider performance implications, especially for data-intensive operations
- For context values, establish consistent conventions for keys and value types
- Middleware can transform or enrich errors for better diagnostics
- Consider testing strategies for middleware components

### 📝 EXAMPLE

```go
func Logging[R any, S any](actionHandler core.ActionHandler[R, S], indentation int) core.ActionHandler[R, S] {
    return func(ctx context.Context, request R) (*S, error) {
        // Before processing
        bytes, err := json.Marshal(request)
        if err != nil {
            return nil, err
        }
        PrintWithIndentation(fmt.Sprintf(">>> REQUEST          %s\n", string(bytes)), indentation)

        // Process
        response, err := actionHandler(ctx, request)
        
        // After processing
        if err != nil {
            PrintWithIndentation(fmt.Sprintf(">>> RESPONSE ERROR  %s\n\n", err.Error()), indentation)
            PrintLine(indentation)
            return nil, err
        }

        bytes, err = json.Marshal(response)
        if err != nil {
            return nil, err
        }
        PrintWithIndentation(fmt.Sprintf(">>> RESPONSE SUCCESS %s\n\n", string(bytes)), indentation)
        PrintLine(indentation)

        return response, nil
    }
}