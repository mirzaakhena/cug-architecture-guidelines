## Rules for WIRING Implementation

### ✅ REQUIRED
- Follow the component setup pattern (gateway → usecase → middleware → controller)
- Apply middleware in the correct order
- Use variable names that reflect component roles with middleware
- Include all dependencies required by usecases
- Add controllers to apiPrinter for documentation

### ❌ FORBIDDEN
- NEVER skip required middleware (logging, transaction)
- NEVER connect controllers directly to usecases without middleware
- NEVER create dependency cycles

### 💡 IMPORTANT TO REMEMBER
- Wiring is where all components are integrated
- Middleware order is critical (e.g., logging → timing → transaction → retry)
- Wiring can be organized by domain or functional module
- Use descriptive variable names for usecases with middleware
- Consider how to handle environment-specific configuration
- Implement strategies for conditional middleware application
- Plan initialization order for multiple domains/modules
- Handle potential errors during the wiring process

### 📝 EXAMPLE

```go
func SetupUserManagement(apiPrinter *utility.ApiPrinter, Mux *http.ServeMux, db *gorm.DB) {
    // Initialize gateways
    findUserByEmail := gateway.ImplFindUserByEmail(db)
    generateUUID := gateway.ImplGenerateUUID()
    getCurrentTime := gateway.ImplGetCurrentTime()
    createUserDB := gateway.ImplCreateUserDB(db)

    // Initialize usecase with middlewares
    createUserUsecase := usecase.ImplCreateUser(
        findUserByEmail,
        generateUUID,
        getCurrentTime,
        createUserDB,
    )

    // Apply middlewares to the usecase
    // 1. Apply logging middleware
    createUserWithLogging := middleware.Logging(createUserUsecase, 0)
    
    // 2. Apply timing middleware to measure performance
    createUserWithTiming := middleware.Timing(createUserWithLogging, "CreateUser")
    
    // 3. Apply transaction middleware for database operations
    createUserWithTransaction := middleware.TransactionMiddleware(createUserWithTiming, db)
    
    // 4. Apply retry middleware for resilience
    createUserWithRetry := middleware.Retry(createUserWithTransaction, 3)

    // Register controller with the final wrapped usecase
    apiPrinter.Add(controller.CreateUserController(Mux, createUserWithRetry))
}