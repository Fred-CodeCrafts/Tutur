# OOP Patterns — Bahasa Daerah Learning Platform (Go Backend)

Go is not a classical OOP language, but it achieves the same goals through interfaces, struct embedding, and composition. This document maps the patterns used in this codebase to their OOP equivalents, with real code references.

---

## 1. Layered Architecture (Handler → Service → Repository)

Every feature module follows the same three-layer structure. Each layer only knows about the layer directly below it via an interface — never the concrete type.

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────────────────┐
│  Handler  (pkg boundary: HTTP concerns only)         │
│  - Decode JSON, validate input, write response       │
│  - Calls Service interface                           │
└──────────────────────┬──────────────────────────────┘
                       │ Service interface
                       ▼
┌─────────────────────────────────────────────────────┐
│  Service  (business logic)                           │
│  - Orchestrates rules, guards, transformations       │
│  - Calls Repository interface                        │
└──────────────────────┬──────────────────────────────┘
                       │ Repository interface
                       ▼
┌─────────────────────────────────────────────────────┐
│  Repository  (data access)                           │
│  - SQL queries, transactions, error mapping          │
│  - Depends only on domain types                      │
└─────────────────────────────────────────────────────┘
```

**Example — auth module:**

```go
// Handler holds only the Service interface, never the concrete service
type Handler struct {
    svc Service          // interface, not *service
}

// Service holds only the Repository interface
type service struct {
    repo Repository      // interface, not *repository
}

// Repository is the concrete struct, hidden behind the interface
type repository struct {
    pool *pgxpool.Pool
}
```

This is the **Dependency Inversion Principle** applied consistently: high-level modules (Handler, Service) depend on abstractions (interfaces), not on concrete implementations.

---

## 2. Interface-Based Polymorphism

Go interfaces are implicit — any type that implements the method set satisfies the interface. This is used in two ways:

### 2a. Service and Repository Interfaces

Each module defines its own interface. The constructor returns the interface, not the concrete type. Callers never see the struct.

```go
// auth/service.go
type Service interface {
    Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
    Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
    UpgradeRole(ctx context.Context, userID uuid.UUID) (*AuthResponse, error)
}

// Returns the interface — caller cannot access unexported fields
func NewService(repo Repository) Service {
    return &service{repo: repo}
}
```

```go
// auth/repository.go
type Repository interface {
    CreateUser(ctx context.Context, user *domain.User) error
    GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
    GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
    UpdateUserRole(ctx context.Context, userID uuid.UUID, role domain.Role) error
}

func NewRepository(pool *pgxpool.Pool) Repository {
    return &repository{pool: pool}
}
```

**Why it matters:** In tests, you swap the real `repository` for a mock that also satisfies `Repository`. The service code is unchanged.

### 2b. AudioStorageService Interface (planned — `internal/storage`)

```go
// Defined in design.md, to be implemented in internal/storage
type AudioStorageService interface {
    Upload(ctx context.Context, fileBytes []byte, mimeType string) (url string, err error)
    GenerateSignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
    Delete(ctx context.Context, objectKey string) error
}
```

Any S3-compatible backend (Cloudflare R2, Supabase Storage, MinIO) can be plugged in by implementing this interface. The phrase submission handler never knows which provider is in use.

---

## 3. Constructor Pattern (Factory Method)

Every struct is created through a `New*` constructor that returns an interface. Direct struct instantiation is never done outside the package.

```go
// auth
authRepo    := auth.NewRepository(pool)   // returns auth.Repository
authSvc     := auth.NewService(authRepo)  // returns auth.Service
authHandler := auth.NewHandler(authSvc)   // returns *auth.Handler

// validation
validationRepo    := validation.NewRepository(pool)
validationSvc     := validation.NewService(validationRepo)
validationHandler := validation.NewHandler(validationSvc)
```

This is the **Factory Method** pattern. The constructor owns the wiring logic and enforces that dependencies are always provided. It also hides the concrete type from the caller.

---

## 4. Strategy Pattern — Middleware Chain

The middleware stack is a chain of `func(http.Handler) http.Handler` functions. Each middleware wraps the next handler, forming a pipeline. You can swap or reorder strategies without touching the handlers.

```go
// pkg/middleware/auth.go

// Authenticate is a strategy: validates JWT, injects identity into context
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... validate token ...
            next.ServeHTTP(w, r.WithContext(ctx))  // delegate to next
        })
    }
}

// RequireRole is a separate strategy: checks role from context
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... check role ...
            next.ServeHTTP(w, r)
        })
    }
}
```

Applied in handlers:

```go
// Different role strategies applied to different route groups
r.Group(func(r chi.Router) {
    r.Use(middleware.Authenticate(jwtSecret))
    r.Use(middleware.RequireRole(domain.RoleContributor, domain.RoleAdmin))
    r.Post("/", h.SubmitPhrase)
})

r.Group(func(r chi.Router) {
    r.Use(middleware.Authenticate(jwtSecret))
    r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))
    r.Get("/{id}", h.GetPhraseByID)
})
```

---

## 5. Facade Pattern — Response Package

`pkg/response` is a facade over `net/http` response writing. It hides the repetitive boilerplate of setting headers, status codes, and encoding JSON behind a clean, named API.

```go
// pkg/response/response.go

// Low-level primitive
func JSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

// Semantic facades — callers use these, never JSON() directly
func BadRequest(w http.ResponseWriter, code, message string) { ... }
func Unauthorized(w http.ResponseWriter)                     { ... }
func Forbidden(w http.ResponseWriter, message string)        { ... }
func NotFound(w http.ResponseWriter, message string)         { ... }
func Conflict(w http.ResponseWriter, code, message string)   { ... }
func UnprocessableEntity(w http.ResponseWriter, ...)         { ... }
func InternalServerError(w http.ResponseWriter)              { ... }
```

Handlers call `response.Conflict(w, "DUPLICATE_EMAIL", "...")` instead of manually setting status 409 and encoding JSON. The error envelope shape is enforced in one place.

---

## 6. Builder Pattern — Validator

`pkg/validator` implements a fluent builder that accumulates field-level errors before deciding whether the input is valid.

```go
// pkg/validator/validator.go

v := validator.New()                          // create builder

v.Check(req.Name != "", "name", "required")   // accumulate checks
v.Check(len(req.Password) >= 8, "password", "min 8 chars")
v.Check(emailRegex.MatchString(req.Email), "email", "invalid format")

if !v.Valid() {                               // inspect result
    err := v.Err()                            // build final error
    response.BadRequest(w, "VALIDATION_ERROR", err.Error())
    return
}
```

The `Validator` struct collects all errors before returning. This means a single request gets all validation failures at once, not just the first one. The `Err()` method builds the final `*ValidationError` value only when needed.

---

## 7. Value Object Pattern — Typed Enums

Domain concepts that have a fixed set of valid values are expressed as typed string constants rather than raw strings. This prevents invalid values from being passed around silently.

```go
// internal/domain/models.go

type Role string
const (
    RoleLearner     Role = "learner"
    RoleContributor Role = "contributor"
    RoleAdmin       Role = "admin"
)

type PhraseStatus string
const (
    StatusPending  PhraseStatus = "pending"
    StatusApproved PhraseStatus = "approved"
    StatusRejected PhraseStatus = "rejected"
    StatusFlagged  PhraseStatus = "flagged"
    StatusAIFailed PhraseStatus = "ai_failed"
)

type VoteType string
const (
    VoteUpvote   VoteType = "upvote"
    VoteDownvote VoteType = "downvote"
)

type FlagReason string
const (
    FlagInaccurateTranslation FlagReason = "inaccurate_translation"
    FlagInappropriateContent  FlagReason = "inappropriate_content"
    FlagDuplicate             FlagReason = "duplicate"
)
```

The compiler enforces type safety: you cannot pass a `Role` where a `PhraseStatus` is expected, even though both are `string` underneath. Validation functions check against the known constants:

```go
func validateVoteType(vt domain.VoteType) error {
    if vt != domain.VoteUpvote && vt != domain.VoteDownvote {
        return ErrInvalidVoteType
    }
    return nil
}
```

---

## 8. Sentinel Error Pattern (Error as Value Object)

Errors are declared as package-level variables and compared with `errors.Is`. This is Go's equivalent of typed exception classes — each error is a distinct, comparable value.

```go
// auth/repository.go
var ErrDuplicateEmail = errors.New("duplicate email")
var ErrNotFound       = errors.New("user not found")

// auth/service.go
var ErrInvalidCredentials  = errors.New("invalid credentials")
var ErrRoleAlreadyUpgraded = errors.New("role already upgraded")

// validation/service.go
var ErrPhraseNotFound = errors.New("phrase not found")
var ErrDuplicateVote  = errors.New("duplicate vote")
var ErrSelfVote       = errors.New("self vote not allowed")
```

Errors are wrapped with context as they propagate up:

```go
// repository wraps low-level DB error
return fmt.Errorf("create user: %w", err)

// service checks for specific sentinel, re-wraps otherwise
if errors.Is(err, ErrNotFound) {
    return nil, ErrInvalidCredentials  // translate to service-level error
}
return nil, fmt.Errorf("get user: %w", err)

// handler checks for service-level sentinel
if errors.Is(err, ErrInvalidCredentials) {
    response.Unauthorized(w)
    return
}
```

The `%w` verb preserves the error chain so `errors.Is` works at any depth.

---

## 9. Domain Model Pattern — Shared Entity Package

All domain entities live in `internal/domain/models.go`. This is a single source of truth for data shapes. No layer defines its own version of `User` or `Phrase`.

```go
// internal/domain/models.go — shared by all layers

type User struct {
    ID           uuid.UUID `json:"id"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`   // never serialised to JSON
    Role         Role      `json:"role"`
    IsActive     bool      `json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type Phrase struct {
    // ... all fields including denormalised vote counts
    Words           []Word           `json:"words,omitempty"`
    CulturalContext *CulturalContext `json:"cultural_context,omitempty"`
}
```

Struct tags control JSON serialisation. `json:"-"` on `PasswordHash` ensures it is never leaked in a response, regardless of which handler serialises the struct.

---

## 10. Composition over Inheritance — Engine inside Service

Go has no inheritance. Complex behaviour is built by composing structs. The `validation.service` embeds an `Engine` to separate threshold logic from orchestration logic.

```go
// validation/service.go
type service struct {
    repo   Repository
    engine *Engine      // composed, not inherited
}

func NewService(repo Repository) Service {
    return &service{
        repo:   repo,
        engine: NewEngine(repo),  // Engine gets the same repo
    }
}
```

```go
// validation/engine.go — focused solely on threshold transitions
type Engine struct {
    repo Repository
}

func (e *Engine) CheckPhraseThresholds(ctx context.Context, phraseID uuid.UUID, upvotes, downvotes int) error {
    switch {
    case upvotes >= upvoteApproveThreshold:
        return e.repo.UpdatePhraseStatus(ctx, phraseID, domain.StatusApproved)
    case downvotes >= downvoteRejectThreshold:
        return e.repo.UpdatePhraseStatus(ctx, phraseID, domain.StatusRejected)
    }
    return nil
}
```

`service` handles orchestration (self-vote check, insert, call engine). `Engine` handles only the threshold decision. Neither knows about HTTP.

---

## 11. Context Propagation — Implicit Dependency Injection

`context.Context` is threaded through every function call. The middleware layer injects identity into the context; downstream code reads it without needing the HTTP request.

```go
// middleware injects
ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
ctx  = context.WithValue(ctx, contextKeyRole, role)
next.ServeHTTP(w, r.WithContext(ctx))

// handler reads
userID, ok := middleware.UserIDFromContext(r.Context())

// service receives context but doesn't know about HTTP
func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
    // passes ctx to repository
    if err := s.repo.CreateUser(ctx, user); err != nil { ... }
}

// repository uses context for DB cancellation and timeout
err := r.pool.QueryRow(ctx, query, ...).Scan(...)
```

This is **implicit dependency injection** via context — identity flows from middleware to handler to service to repository without being passed as explicit parameters.

---

## Pattern Summary

| Pattern | Where used | Go mechanism |
|---|---|---|
| Layered Architecture | All modules | Package separation, interface boundaries |
| Interface Polymorphism | Service, Repository, Storage | `interface` + implicit satisfaction |
| Factory Method | All `New*` constructors | Constructor functions returning interfaces |
| Strategy | Middleware chain | `func(http.Handler) http.Handler` |
| Facade | `pkg/response` | Named wrapper functions |
| Builder | `pkg/validator` | Accumulator struct + `Err()` finaliser |
| Value Object | Domain enums (`Role`, `PhraseStatus`, etc.) | Typed string constants |
| Sentinel Error | All error variables | `errors.New` + `errors.Is` + `%w` wrapping |
| Domain Model | `internal/domain/models.go` | Shared struct package |
| Composition | `validation.service` + `Engine` | Struct field embedding |
| Context Propagation | Middleware → Handler → Service → Repository | `context.Context` threading |
