## Rules for HTTP WRITE CONTROLLER Implementation

### ✅ REQUIRED
- Use functions for controllers, not structs/interfaces
- Accept usecases as parameters
- Use `utils.HandleUsecase` for request parsing, validation, and error handling
- Follow naming convention: `[UsecaseName]Controller`
- Return `utils.APIData` for automatic documentation
- Handle path parameters via `r.PathValue("paramName")`
- Extract query parameters using `utils.GetQueryString()` and similar methods
- Use JSONTag from usecase request for body parsing
- Use appropriate HTTP methods (POST, PUT, PATCH, DELETE) for write operations
- Set appropriate HTTP status codes (201 for creation, 200/204 for updates/deletes)

### ❌ FORBIDDEN
- NEVER put business logic in controllers
- NEVER access databases or infrastructure directly (except through gateways)
- NEVER duplicate request/response structs
- NEVER call more than one usecase from a single controller
- NEVER use hardcoded URLs or routes (use constants or configuration)

### 💡 IMPORTANT TO REMEMBER
- HTTP controllers are concerned with HTTP-specific concerns (routes, methods, headers, status codes)
- Controllers function as adapters between HTTP and usecases
- Customize error responses based on error types where appropriate
- Include CORS handling where needed for browser clients
- Consider versioning strategy for APIs (path, header, or content negotiation)
- Use `apiPrinter` to document all endpoints systematically
- For file uploads, use multipart form handling

### 📝 HTTP CREATE CONTROLLER EXAMPLE

```go
func CreateUserController(Mux *http.ServeMux, u usecase.CreateUser) utils.APIData {
    apiData := utils.APIData{
        Method:      http.MethodPost,
        Url:         "/api/users",
        Body:        usecase.CreateUserReq{}, // Using the same request struct as the usecase
        Summary:     "Create a new user",
        Description: "Register a new user account with email and name",
        Tag:         "User Management",
        Examples: []utils.ExampleResponse{
            {
                StatusCode: 200,
                Content: map[string]any{
                    "status": "success", 
                    "data": map[string]string{
                        "userId": "123e4567-e89b-12d3-a456-426614174000",
                    },
                },
            },
            {
                StatusCode: 400,
                Content: map[string]any{
                    "status": "failed", 
                    "error": "email already registered",
                },
            },
        },
    }

    handler := func(w http.ResponseWriter, r *http.Request) {
        // Parse request body
        req, ok := utils.ParseJSON[usecase.CreateUserReq](w, r)
        if !ok {
            // ParseJSON already handled the error response
            return
        }

        // Handle the usecase call with the request body
        utils.HandleUsecase(r.Context(), w, u, req)
    }

    Mux.HandleFunc(apiData.Method+" "+apiData.Url, handler)

    return apiData
}
```