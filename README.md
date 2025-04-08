# Clean Architecture Convention Guide for AI

## Project Folder 

```
.
├── go.work
├── go.work.sum
└── project
   ├── controller
   ├── core
   │   └── core.go
   ├── gateway
   │   └── undeterministic.go
   ├── go.mod
   ├── go.sum
   ├── main.go
   ├── middleware
   │   ├── common.go
   │   └── transaction.go
   ├── model
   ├── usecase
   ├── utils
   │   ├── controller.go
   │   └── printer_http_api.go
   └── wiring
      └── setup.go
```


## Main Principles

The code must follow Clean Architecture with some aspects of Hexagonal Architecture:

1. **Layer Separation**:
   - **Controller**: Publishes usecases to the outside world through various protocols (HTTP, MQTT, gRPC, event handler, CLI, scheduler, etc.)
   - **Usecase**: Contains pure business logic, orchestrates gateways, has no infrastructure dependencies
   - **Gateway**: Handles communication with external infrastructure and non-deterministic functions
   - **Middleware**: Decorators for usecases that handle cross-cutting concerns
   - **Wiring**: Dependency injection and component composition

2. **Dependency Flow**:
   - Controller → Usecase → Gateway
   - Dependencies only flow inward (inner layers must not know about outer layers)
   - Usecases may have dependencies on multiple Gateways
   - Gateways must not have dependencies on other Gateways
   - A Controller must not call more than one Usecase
   - Controllers may directly call Gateways in certain cases (simple CRUD operations)

3. **Common Types**:
   - Use `core.ActionHandler[Request, Response]` as the standard function type for Usecases and Gateways
   - All functions accept context.Context as the first parameter
   - This type supports functional composition and avoids "fat interface" problems
   - Request and Response use structs with consistent naming (GatewayNameReq/Res, UsecaseNameReq/Res)

## Correct Development Sequence

1. **Start with Usecase**: 
   - Define the usecase interface and algorithm
   - Identify required gateways (still as empty interfaces)
   - Focus on business flow and rules without being tied to infrastructure implementation

2. **Implement Gateways**:
   - Implement the gateways identified in the previous step
   - Prioritize gateways that could potentially be reused by other usecases
   - Basic CRUD gateways can be implemented first as a foundation

3. **Create Controllers**:
   - Implement controllers that expose usecases (or gateways directly for simple CRUD operations)
   - Adapt to the protocol being used (HTTP, MQTT, gRPC, etc.)

4. **Setup Wiring**:
   - Connect all components with dependency injection
   - Apply middleware as needed (not all gateways or usecases must have middleware)

5. **Unit Testing**:
   - Create unit tests for usecases with gateway mocking
   - Verify usecase behavior in various scenarios