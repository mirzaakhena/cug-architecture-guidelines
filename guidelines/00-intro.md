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
   - All functions accept `context.Context` as the first parameter
   - This type supports functional composition and avoids "fat interface" problems
   - Request and Response use structs with consistent naming (GatewayNameReq/Res, UsecaseNameReq/Res)

## Correct Development Sequence

The development process is divided into three main phases, each with specific focus areas and documentation resources. Follow this sequence for optimal implementation:

### Phase 1: Usecase Planning & Gateway Implementation
**Documentation Resources: [./guidelines/01-usecase.md], [./guidelines/02-gateway.md]**

1. **Initial Usecase Planning**:
   - Define the usecase's purpose and basic requirements
   - Draft the usecase request/response structures
   - Identify all gateway dependencies needed for the usecase
   - Focus on business flow and rules without implementation details

2. **Scan Existing Models**:
   - Review all domain models in the project's `model/` directory that relate to the usecase
   - Determine which existing models can be reused or extended for the new usecase
   - Identify any new models that need to be created to support the business requirements
   - Note relationships between models that might affect the implementation
   - Check for validation rules and business constraints already defined in existing models

3. **Scan Existing Gateways**:
   - Review existing gateway interfaces in the project's `gateway/` directory against the identified needs
   - Check for reusable gateway implementations by examining all `.go` files in the `gateway/` folder
   - Determine which existing gateways can be reused without modification
   - Document which new gateways need to be created or which existing ones need to be extended
   - Note any gateways that might need refactoring to accommodate new requirements

4. **Implement Missing Gateways**:
   - Create any gateways identified as missing but required
   - Prioritize gateways that could potentially be reused by other usecases
   - Implement basic CRUD gateways first as a foundation where needed
   - Ensure proper error handling and infrastructure abstraction

### Phase 2: Usecase & Controller Implementation
**Documentation Resources: [./guidelines/01-usecase.md], [./guidelines/03-controller-http-read.md], [./guidelines/03-controller-http-write.md], [./guidelines/03-controller-subscriber.md], [./guidelines/03-controller-scheduler.md]**

5. **Implement Complete Usecase**:
   - With all gateways now available, implement the full usecase logic
   - Focus on business rules and orchestration of the gateways
   - Handle all error cases and edge conditions
   - Ensure pure business logic remains separate from infrastructure concerns

6. **Create Controllers**:
   - Implement controllers that expose usecases (or gateways directly for simple CRUD operations)
   - Choose the appropriate controller type based on exposure requirements:
     - HTTP Read: `./guidelines/03-controller-http-read.md` (for GET endpoints)
     - HTTP Write: `./guidelines/03-controller-http-write.md` (for POST/PUT/DELETE endpoints)
     - Message Queue: `./guidelines/03-controller-subscriber.md` (for async messaging)
     - Scheduler: `./guidelines/03-controller-scheduler.md` (for time-based operations)
   - Adapt to the protocol being used (HTTP, MQTT, gRPC, etc.)
   - Handle protocol-specific parameter extraction and response formatting
   - Document endpoints for API documentation

### Phase 3: Wiring & Testing
**Documentation Resources: [./guidelines/04-middleware.md], [./guidelines/05-wiring.md]**

7. **Setup Wiring**:
   - Connect all components with proper dependency injection
   - Apply middleware as needed (not all gateways or usecases must have middleware)
   - Use the patterns documented in `./guidelines/04-middleware.md` for cross-cutting concerns
   - Register controllers with their respective service handlers
   - Ensure proper initialization order for all components
   - Follow the structure outlined in `./guidelines/05-wiring.md`

8. **Unit Testing**:
   - Create unit tests for usecases with gateway mocking
   - Verify usecase behavior in various scenarios
   - Test error paths and edge cases
   - Ensure coverage of critical business logic

9. **Integration Testing**:
   - Test the flow through controllers, usecases, and gateways working together
   - Verify middleware behavior and proper context propagation
   - Test error handling across component boundaries
   - Validate that all components interact as expected in the wired application