## Rules for HTTP READ CONTROLLER Implementation

### ✅ REQUIRED
- Use functions for controllers, not structs/interfaces
- Accept usecases as parameters
- Use `utils.HandleUsecase` for request parsing, validation, and error handling
- Follow naming convention: `[UsecaseName]Controller`
- Return `utils.APIData` for automatic documentation
- Handle path parameters via `r.PathValue("paramName")`
- Extract query parameters using `utils.GetQueryString()` and similar methods
- Always use HTTP GET method for read operations
- Set appropriate HTTP status codes (200 for success, 404 for not found)

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

### 📝 HTTP GET CONTROLLER EXAMPLE

```go
func GetUserController(Mux *http.ServeMux, u usecase.GetUser) utils.APIData {
    apiData := utils.APIData{
        Method:  http.MethodGet,
        Url:     "/api/users/{userId}",
        Summary: "Get user details",
        Tag:     "User Management",
        QueryParams: []utils.QueryParam{
            {
                Name:        "include",
                Type:        "string",
                Description: "Comma-separated list of related resources to include",
                Required:    false,
            },
        },
    }

    handler := func(w http.ResponseWriter, r *http.Request) {
        // Extract path parameter
        userId := r.PathValue("userId")
        
        // Extract query parameter
        include := utils.GetQueryString(r, "include", "")
        
        // Create request for usecase
        req := usecase.GetUserReq{
            UserID:  userId,
            Include: include,
        }
        
        // Handle the usecase call
        utils.HandleUsecase(r.Context(), w, u, req)
    }

    Mux.HandleFunc(apiData.Method+" "+apiData.Url, handler)

    return apiData
}
```
