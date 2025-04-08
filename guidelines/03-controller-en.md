## Rules for CONTROLLER Implementation

### ✅ REQUIRED
- Use functions for controllers, not structs/interfaces
- Accept usecases or gateways as parameters
- Use utils.HandleUsecase for request parsing, validation, and error handling
- Follow naming convention: `[UsecaseName]Controller`
- Return utils.APIData for documentation
- Connect controllers to appropriate access mechanisms (HTTP, MQTT, scheduler, event handler, etc.)
- Use JSONTag from usecase request for body parsing

### ❌ FORBIDDEN
- NEVER put business logic in controllers
- NEVER access databases or infrastructure directly (except through gateways)
- NEVER duplicate request/response structs
- NEVER call more than one usecase from a single controller

### 💡 IMPORTANT TO REMEMBER
- Controllers may call gateways directly for simple CRUD operations
- Controllers function as adapters between the outside world and usecases/gateways
- Controllers handle various communication protocols, not just HTTP
- Controllers can transform data formats according to protocol needs
- Consider handling authentication/authorization at the controller level
- Customize error responses based on error types where appropriate
- For complex input transformations, consider helper functions

## Best Practices
- **One Usecase Per Controller**: Each controller should call only one usecase
- **Parameter Extraction**: Use utils functions consistently for all parameter types
- **Protocol-Specific Logic**: Keep protocol-specific logic in controller, business logic in usecase
- **API Documentation**: Use APIData to document endpoints for auto-generated documentation
- **Descriptive Parameters**: Use meaningful names for path and query parameters

### 📝 HTTP CONTROLLER EXAMPLE

```go
func CreateUserController(Mux *http.ServeMux, u usecase.CreateUser) utils.APIData {

    apiData := utils.APIData{
        Method:  http.MethodPost,
        Url:     "/api/users/{id}",
        Body:    usecase.CreateUserReq{}, // Using the same request struct as the usecase
        Summary: "Create a new user",
        Tag:     "User Management",
    }

    handler := func(w http.ResponseWriter, r *http.Request) {

        // This is how to get path parameter
        id := r.PathValue("id")

        // This is how to get query parameters for additional options
        name := utils.GetQueryString(r, "name", "")

        // This is how to parse request body
        req, ok := utility.ParseJSON[usecase.CreateUserReq](w, r)
        if !ok {
            // ParseJSON already handled the error response
            return
        }

        // some value need to be assigned manually like this
        req.ID = id
        req.Name = name

        // Handle the usecase call with the request body
        utils.HandleUsecase(r.Context(), w, u, req)
    }

    Mux.HandleFunc(apiData.GetMethodUrl(), handler)

    return apiData
}
```

### 📝 MQTT CONTROLLER EXAMPLE

```go
func HandleAgentHeartbeatController(client mqtt.Client, u usecase.ProcessAgentHeartbeat) {
    topic := "nms/agents/+/heartbeat"
    
    callback := func(client mqtt.Client, msg mqtt.Message) {
        // Extract agent ID from topic
        parts := strings.Split(msg.Topic(), "/")
        agentID := parts[2]
        
        // Parse payload
        var heartbeatData usecase.ProcessAgentHeartbeatReq
        if err := json.Unmarshal(msg.Payload(), &heartbeatData); err != nil {
            log.Printf("Error parsing heartbeat payload: %v", err)
            return
        }
        
        // Set agent ID in request
        heartbeatData.AgentID = agentID
        
        // Execute usecase
        _, err := u(context.Background(), heartbeatData)
        if err != nil {
            log.Printf("Error processing heartbeat: %v", err)
            return
        }
    }
    
    token := client.Subscribe(topic, 1, callback)
    token.Wait()
}
```

### 📝 SCHEDULER CONTROLLER EXAMPLE

```go
func SetupCertificateExpiryCheckJob(scheduler *scheduler.Scheduler, u usecase.CheckCertificateExpiry) {
    job := func() {
        _, err := u(context.Background(), usecase.CheckCertificateExpiryReq{})
        if err != nil {
            log.Printf("Certificate expiry check failed: %v", err)
        }
    }
    
    // Schedule job to run daily at 1 AM
    scheduler.ScheduleDaily("certificate-expiry-check", "1:00", job)
}
```
