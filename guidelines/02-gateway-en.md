## Rules for GATEWAY Implementation

### Gateway Naming Conventions

When implementing gateways in our clean architecture, follow these naming patterns for consistency and clarity:

#### Standard Pattern: `[Object][Action]`

Use the format `[Object][Action]` for all gateways where:
- `Object` is the domain entity or resource being operated on
- `Action` is the operation being performed

Examples:
- `UserSave` - Saves a user to the database
- `UserFindOne` - Retrieves a single user
- `UserFindMany` - Retrieves multiple users
- `MessagePublish` - Publishes a message to a queue
- `AgentActivate` - Activates an agent

This naming convention provides natural grouping in file systems and code editors, making related gateways easy to locate.

#### System Operations: `System[Action]`

For non-deterministic utility operations without a clear domain entity, use `System` as the object:

- `SystemGenerateUUID` - Generates a UUID
- `SystemGenerateRandomString` - Generates a random string
- `SystemGetCurrentTime` - Retrieves the current time

This maintains our naming pattern consistency while providing a logical home for operations that don't belong to a specific domain entity.

### ✅ REQUIRED
- Create **each gateway for ONE specific task** (Single Responsibility)
- Structure gateways as **functions rather than structs/interfaces**
- Use the `core.ActionHandler[Req, Res]` type for consistency
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

### 💡 IMPORTANT TO REMEMBER
- Gateways are responsible for all non-deterministic operations
- Gateways should be independent and reusable by different usecases
- Gateways can access infrastructure like databases, APIs, files, or external functions
- Gateways are the appropriate place for time.Now(), UUID generation, random number generation, etc.
- Consider categories of gateways (database, external API, file system) and their specific requirements
- Use appropriate error types for different scenarios (e.g., distinguishing between not found vs. server errors)

### 📝 EXAMPLE

```go
// UserFindByEmail gateway
type UserFindByEmailReq struct {
    Email string
}

type UserFindByEmailRes struct {
    Exists bool
    User   *User // Only populated if user exists
}

type UserFindByEmail = core.ActionHandler[UserFindByEmailReq, UserFindByEmailRes]

func ImplUserFindByEmail(db *gorm.DB) UserFindByEmail {
    return func(ctx context.Context, req UserFindByEmailReq) (*UserFindByEmailRes, error) {

        // Get transaction from context if available
        dbCtx := middleware.GetDBFromContext(ctx, db)
        
        var user User
        result := dbCtx.Where("email = ?", req.Email).First(&user)
        
        if result.Error != nil {
            if result.Error == gorm.ErrRecordNotFound {
                return &UserFindByEmailRes{
                    Exists: false,
                    User:   nil,
                }, nil
            }
            return nil, fmt.Errorf("database error: %w", result.Error)
        }
        
        return &UserFindByEmailRes{
            Exists: true,
            User:   &user,
        }, nil
    }
}
```