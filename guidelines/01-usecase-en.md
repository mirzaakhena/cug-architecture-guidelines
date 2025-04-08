## Rules for USECASE Implementation

### ✅ REQUIRED
- Implement usecases as **pure functions** without side effects
- Place **all business logic** in usecases
- Orchestrate gateways to perform infrastructure operations
- Define clear request & response structs with consistent naming: `[UsecaseName]Req` and `[UsecaseName]Res`
- Use the `core.ActionHandler[Req, Res]` type
- Implement using closure pattern: `func Impl[UsecaseName](gateways) [UsecaseName] { return func(ctx, req) {} }`
- Validate all inputs before calling gateways
- Wrap errors with contextual information: `fmt.Errorf("failed to...: %w", err)`
- Always return pointers: `(*Response, error)` rather than `(Response, error)`
- Usecase function parameters should only be gateway dependencies

### ❌ FORBIDDEN
- Never implement non-deterministic functions (time.Now(), uuid.New()) in usecases
- Never access databases or APIs directly
- Never ignore errors returned from gateways
- Never have infrastructure dependencies of any kind
- Never call other usecases from within a usecase (only orchestrate gateways)
- Never use global state or singletons

### 💡 IMPORTANT TO REMEMBER
- Usecases may exist without any gateway dependencies (pure data transformation)
- Complexity in usecases reflects business process complexity
- Consider breaking complex usecases into private helper functions or multiple smaller usecases
- Request structs can have a `Validate()` method for internal validation
- Different error types can provide more context about failure modes
- Use domain-specific error types to make error handling more informative
- Consider how context values are propagated through to gateways for tracing and cancellation
- When returning errors, wrap them with sufficient context for debugging and user feedback

### 📝 EXAMPLE

```go
// CreateUser usecase
type CreateUserReq struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

// Input Validation
func (u CreateUserReq) Validate() error {
    if u.Email == "" {
        return errors.New("email is required")
    }
    if u.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

type CreateUserRes struct {
    UserID string `json:"userId"`
}

type CreateUser = core.ActionHandler[CreateUserReq, CreateUserRes]

func ImplCreateUser(
    findUserByEmail gateway.FindUserByEmail,
    generateUUID gateway.GenerateUUID,
    getCurrentTime gateway.GetCurrentTime,
    createUserDB gateway.CreateUserDB,
) CreateUser {
    return func(ctx context.Context, req CreateUserReq) (*CreateUserRes, error) {

        // 1. Validate input
        if err := req.Validate(); err != nil {
            return nil, err
        }

        // 2. Check if email already exists
        existingUserRes, err := findUserByEmail(ctx, gateway.FindUserByEmailReq{
            Email: req.Email,
        })
        if err != nil {
            return nil, err
        }
        if existingUserRes.Exists {
            return nil, errors.New("email already registered")
        }

        // 3. Generate UUID for new user
        uuidRes, err := generateUUID(ctx, gateway.GenerateUUIDReq{})
        if err != nil {
            return nil, err
        }

        // 4. Get current time
        timeRes, err := getCurrentTime(ctx, gateway.GetCurrentTimeReq{})
        if err != nil {
            return nil, err
        }

        // 5. Create user in database
        createRes, err := createUserDB(ctx, gateway.CreateUserDBReq{
            ID:        uuidRes.Value,
            Email:     req.Email,
            Name:      req.Name,
            CreatedAt: timeRes.Now,
        })
        if err != nil {
            return nil, err
        }

        // 6. Return result
        return &CreateUserRes{
            UserID: createRes.UserID,
        }, nil
    }
}