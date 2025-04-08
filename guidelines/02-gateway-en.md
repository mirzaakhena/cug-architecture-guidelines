## Rules for GATEWAY Implementation

### ✅ REQUIRED
- Create **each gateway for ONE specific task** (Single Responsibility)
- Structure gateways as **functions rather than structs/interfaces**
- Use the `core.ActionHandler[Req, Res]` type for consistency
- Define clear request & response structs with naming convention: `[GatewayName]Req` and `[GatewayName]Res`
- Implement using closure pattern: `func Impl[GatewayName](dependencies) [GatewayName] { return func(ctx, req) {} }`
- Include comprehensive comments for each gateway
- Wrap errors with context: `fmt.Errorf("failed to...: %w", err)`
- Always return pointers: `(*Response, error)`
- Use `middleware.GetDBFromContext(ctx, db)` for transaction management
- Gateway function parameters can be of any type needed for infrastructure access

### ❌ FORBIDDEN
- NEVER access usecases from gateways
- NEVER put business logic in gateways
- NEVER return database errors directly without wrapping
- NEVER hardcode values that should be parameters
- NEVER call other gateways from a gateway
- NEVER implement non-deterministic functions (time.Now(), uuid.New()) in usecases

### 💡 IMPORTANT TO REMEMBER
- Gateways are responsible for all non-deterministic operations
- Gateways should be independent and reusable by different usecases
- Gateways can access infrastructure like databases, APIs, files, or external functions
- Gateways are the appropriate place for time.Now(), UUID generation, random number generation, etc.
- Consider categories of gateways (database, external API, file system) and their specific requirements
- Ensure gateways are testable through proper dependency injection
- Use appropriate error types for different scenarios (e.g., distinguishing between not found vs. server errors)

### 📝 EXAMPLE

```go
// FindUserByEmail gateway
type FindUserByEmailReq struct {
    Email string
}

type FindUserByEmailRes struct {
    Exists bool
    User   *User // Only populated if user exists
}

type FindUserByEmail = core.ActionHandler[FindUserByEmailReq, FindUserByEmailRes]

func ImplFindUserByEmail(db *gorm.DB) FindUserByEmail {
    return func(ctx context.Context, req FindUserByEmailReq) (*FindUserByEmailRes, error) {

        // Get transaction from context if available
        dbCtx := middleware.GetDBFromContext(ctx, db)
        
        var user User
        result := dbCtx.Where("email = ?", req.Email).First(&user)
        
        if result.Error != nil {
            if result.Error == gorm.ErrRecordNotFound {
                return &FindUserByEmailRes{
                    Exists: false,
                    User:   nil,
                }, nil
            }
            return nil, fmt.Errorf("database error: %w", result.Error)
        }
        
        return &FindUserByEmailRes{
            Exists: true,
            User:   &user,
        }, nil
    }
}