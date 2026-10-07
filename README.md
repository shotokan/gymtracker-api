# GymTracker API

A serverless REST API for tracking gym workouts, built with Go and AWS CDK.

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (also known as Ports and Adapters), which provides clear separation of concerns and makes the codebase highly testable and maintainable.

### What is Hexagonal Architecture?

Hexagonal Architecture organizes code in layers where the **business logic** (domain) is at the center, completely isolated from external concerns like databases, HTTP, or third-party APIs. All interactions with the outside world happen through well-defined **ports** and **adapters**.

```
┌─────────────────────────────────────────────────────────────────┐
│                         EXTERNAL WORLD                           │
│  (HTTP requests, databases, Google API, AWS services, etc.)     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
         ┌────────────────────────────────────────┐
         │        PRIMARY ADAPTERS                │
         │  (Incoming: HTTP handlers, routers)    │
         └────────────────┬───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │    PRIMARY PORTS      │
              │  (Input interfaces)   │
              │  - Use Cases          │
              └───────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │       DOMAIN LAYER         │
         │  (Business Logic & Rules)  │
         │  - Entities                │
         │  - Domain Errors           │
         │  - Business Rules          │
         └────────────┬───────────────┘
                      │
                      ▼
              ┌───────────────────────┐
              │   SECONDARY PORTS     │
              │  (Output interfaces)  │
              │  - Repositories       │
              │  - External Services  │
              └───────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────────────────┐
         │      SECONDARY ADAPTERS                │
         │  (Outgoing: DynamoDB, bcrypt, Google)  │
         └────────────────────────────────────────┘
```

### Directory Structure Explained

Each module follows this structure:

```
internal/<module>/
├── domain/                    ← Core business logic
│   ├── <entity>.go           ← Business entities/models
│   ├── errors.go             ← Domain-specific errors
│   └── ...
│
├── port/
│   ├── input/                ← PRIMARY PORTS (driving side)
│   │   └── <module>_ports.go ← Interfaces that define use cases
│   │                           (what the application can do)
│   │
│   └── output/               ← SECONDARY PORTS (driven side)
│       ├── repository.go     ← Interfaces for data access
│       └── ...go             ← Interfaces for external services
│
├── application/              ← Use case implementations
│   ├── <use_case>.go        ← Business logic orchestration
│   └── <use_case>_test.go   ← Unit tests
│
└── adapter/
    ├── primary/              ← PRIMARY ADAPTERS (incoming)
    │   ├── router.go        ← Routes requests to handlers
    │   ├── <handler>.go     ← HTTP request handlers
    │   └── helpers.go       ← JSON response helpers
    │
    └── secondary/            ← SECONDARY ADAPTERS (outgoing)
        ├── dynamo/          ← DynamoDB implementations
        ├── bcrypt/          ← Password hashing
        └── google/          ← Google API integration
```

### Layer Responsibilities

#### 1. **Domain** (`domain/`)
- **What**: Core business entities and rules
- **Dependencies**: NONE (pure Go, no external imports)
- **Examples**:
  - `User`, `Plan`, `Session` entities
  - Business rules (e.g., "active plan validation")
  - Domain errors (`ErrNotFound`, `ErrInvalidCreds`)
- **Why isolated**: Business logic should never depend on infrastructure

#### 2. **Ports** (`port/input/` and `port/output/`)
- **What**: Interfaces that define contracts
- **Dependencies**: Only domain layer
- **Types**:
  - **Input Ports** (Primary): Use case interfaces that define what the app can do
    ```go
    type LoginUseCase interface {
        Execute(ctx context.Context, cmd LoginCommand) (LoginResult, error)
    }
    ```
  - **Output Ports** (Secondary): Interfaces for external dependencies
    ```go
    type UserRepository interface {
        FindByEmail(ctx context.Context, email string) (*User, error)
        Save(ctx context.Context, user *User) error
    }
    ```
- **Why interfaces**: Enables dependency inversion and easy testing

#### 3. **Application** (`application/`)
- **What**: Use case implementations (orchestration logic)
- **Dependencies**: Domain + Ports (interfaces only)
- **Responsibilities**:
  - Coordinate between domain and ports
  - Validate input
  - Call domain methods
  - Call repositories/services via ports
- **Example**:
  ```go
  func (uc *loginUseCase) Execute(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
      user, err := uc.userRepo.FindByEmail(ctx, cmd.Email)  // Port
      if err != nil {
          return LoginResult{}, domain.ErrInvalidCreds      // Domain error
      }

      if err := uc.hasher.Compare(user.PasswordHash, cmd.Password); err != nil {
          return LoginResult{}, domain.ErrInvalidCreds
      }

      token, err := uc.jwtGen.Generate(user.ID, user.Email) // Port
      return LoginResult{Token: token}, nil
  }
  ```

#### 4. **Primary Adapters** (`adapter/primary/`)
- **What**: Handle incoming requests (HTTP, CLI, etc.)
- **Dependencies**: Application layer (use cases)
- **Responsibilities**:
  - Parse HTTP requests
  - Route to correct handler
  - Call use cases
  - Format HTTP responses
- **Example**:
  ```go
  type LoginHandler struct {
      uc input.LoginUseCase
  }

  func (h *LoginHandler) Handle(ctx context.Context, req APIGatewayRequest) (Response, error) {
      var body struct {
          Email    string `json:"email"`
          Password string `json:"password"`
      }
      json.Unmarshal(req.Body, &body)

      result, err := h.uc.Execute(ctx, input.LoginCommand{
          Email: body.Email,
          Password: body.Password,
      })

      return jsonResp(200, result)
  }
  ```

#### 5. **Secondary Adapters** (`adapter/secondary/`)
- **What**: Implement output ports with real infrastructure
- **Dependencies**: External libraries (AWS SDK, bcrypt, Google API, etc.)
- **Examples**:
  - `dynamo/user_repository.go` - Implements `UserRepository` using DynamoDB
  - `bcrypt/password_hasher.go` - Implements `PasswordHasher` using bcrypt
  - `google/token_validator.go` - Implements `GoogleTokenValidator` using Google API
- **Why separated**: Easy to swap implementations (e.g., DynamoDB → PostgreSQL)

### Module Examples

#### Auth Module (`internal/auth/`)

```
internal/auth/
├── domain/
│   ├── errors.go              # ErrInvalidCreds, ErrInvalidToken
│   └── google_user.go         # GoogleUser entity
│
├── port/
│   ├── input/
│   │   └── auth_ports.go      # LoginUseCase, GoogleAuthUseCase interfaces
│   │
│   └── output/
│       ├── user_repository.go      # User data access interface
│       ├── password_hasher.go      # Password comparison interface
│       └── google_validator.go     # Google token validation interface
│
├── application/
│   ├── login_use_case.go           # Email/password login logic
│   ├── login_use_case_test.go      # Unit tests with mocks
│   ├── google_auth_use_case.go     # Google OAuth logic
│   └── google_auth_use_case_test.go
│
└── adapter/
    ├── primary/
    │   ├── router.go               # Routes /auth/login and /auth/google
    │   ├── login_handler.go        # Handles login HTTP requests
    │   └── google_auth_handler.go  # Handles Google OAuth requests
    │
    └── secondary/
        ├── bcrypt/password_hasher.go    # bcrypt implementation
        ├── google/token_validator.go    # Google API integration
        └── jwt/token_generator.go       # JWT token generation
```

**Flow Example** (Login):
```
1. API Gateway receives POST /auth/login
   │
2. Lambda (cmd/auth/main.go) starts
   │
3. Router.Handle() → routes to LoginHandler
   │
4. LoginHandler.Handle()
   ├─ Parses JSON request
   └─ Calls LoginUseCase.Execute()
       │
5. LoginUseCase.Execute()
   ├─ userRepo.FindByEmail()     ← Secondary adapter (DynamoDB)
   ├─ hasher.Compare()            ← Secondary adapter (bcrypt)
   └─ jwtGen.Generate()           ← Secondary adapter (JWT)
   │
6. LoginHandler formats JSON response
   │
7. Returns to API Gateway
```

### Why This Architecture?

#### ✅ **Testability**
- Domain and application layers have NO external dependencies
- Easy to mock ports with interfaces
- Fast unit tests (no databases or HTTP servers)

```go
// Example: Testing without real dependencies
func TestLogin(t *testing.T) {
    mockRepo := new(mockUserRepository)
    mockHasher := new(mockPasswordHasher)
    mockJWT := new(mockJWTGenerator)

    useCase := NewLogin(mockRepo, mockHasher, mockJWT)

    mockRepo.On("FindByEmail", "user@example.com").Return(user, nil)
    mockHasher.On("Compare", hash, "password").Return(nil)
    mockJWT.On("Generate", userID, email).Return("token", nil)

    result, err := useCase.Execute(ctx, LoginCommand{...})
    // Test passes without touching DynamoDB or bcrypt
}
```

#### ✅ **Maintainability**
- Clear boundaries: Business logic never mixed with HTTP or DB code
- Easy to find things: "Where's the login logic?" → `application/login_use_case.go`
- Independent layers: Change DB from DynamoDB to Postgres? Only touch `adapter/secondary/`

#### ✅ **Flexibility**
- Swap implementations easily (ports are interfaces)
- Multiple primary adapters (HTTP, CLI, gRPC) can use same use cases
- Multiple secondary adapters (DynamoDB, Postgres, in-memory) for same port

#### ✅ **Business Focus**
- Domain layer contains ONLY business rules
- Infrastructure details hidden behind ports
- Developers focus on "what" not "how"

### Naming Conventions

| Term | Meaning | Example |
|------|---------|---------|
| **Port** | Interface (contract) | `UserRepository`, `LoginUseCase` |
| **Adapter** | Implementation | `DynamoUserRepository`, `LoginHandler` |
| **Primary** | Incoming (drives the app) | HTTP handlers, CLI commands |
| **Secondary** | Outgoing (driven by app) | Database, external APIs |
| **Use Case** | Application service | `LoginUseCase`, `CreatePlanUseCase` |
| **Command** | Input DTO | `LoginCommand`, `CreatePlanCommand` |
| **Result** | Output DTO | `LoginResult`, `PlanResult` |

### Dependency Rule

**Dependencies always point inward:**

```
External World
    ↓
Primary Adapters  →  Primary Ports
                          ↓
                      DOMAIN ← (no dependencies!)
                          ↑
Secondary Adapters ← Secondary Ports
    ↑
External World
```

**Never**:
- ❌ Domain depends on adapters
- ❌ Application depends on adapters
- ❌ Ports import concrete implementations

**Always**:
- ✅ Adapters depend on ports (interfaces)
- ✅ Application depends on ports
- ✅ Domain depends on nothing

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- AWS CLI configured
- Node.js 18+ (for CDK)
- AWS account with appropriate permissions

### Initial Setup

```bash
# 1. Assume deploy role
awsume gymtracker-deploy

# 2. Verify credentials
aws sts get-caller-identity

# 3. Create JWT secret in SSM
make secret

# 4. Bootstrap CDK (first time only)
cd infra && npx cdk bootstrap && cd ..

# 5. Deploy everything
make deploy
```

### Development Workflow

```bash
# 1. Assume the role (expires every hour)
awsume gymtracker-deploy

# 2. Make changes to Go code...

# 3. Run tests
make test

# 4. Run linter
make lint

# 5a. Code-only changes → hotswap (~15s)
make deploy-hotswap

# 5b. Infrastructure changes → full deploy (~3-5 min)
make deploy
```

## 📁 Project Structure

```
.
├── cmd/                      # Lambda entry points
│   ├── auth/                 # Unified auth Lambda (login + Google OAuth)
│   ├── authorizer/           # JWT authorizer
│   ├── user/                 # User operations
│   ├── plan/                 # Training plans
│   └── session/              # Workout sessions
│
├── internal/                 # Private application code
│   ├── auth/                 # Authentication module
│   ├── user/                 # User management module
│   ├── plan/                 # Training plans module
│   ├── session/              # Workout tracking module
│   └── authorizer/           # JWT validation module
│
├── infra/                    # AWS CDK infrastructure
│   ├── lib/
│   │   ├── constructs/       # Reusable CDK constructs
│   │   └── gym-tracker-stack.ts
│   └── bin/
│       └── infra.ts
│
├── docs/                     # Documentation
├── bin/                      # Compiled Lambda binaries (gitignored)
└── makefile                  # Build and deploy commands
```

## 🛠️ Available Commands

```bash
make help            # Show all available commands
make build           # Compile Lambda binaries
make test            # Run tests with race detector
make lint            # Run golangci-lint
make cover           # Generate coverage report
make deploy          # Full deploy (infrastructure + code)
make deploy-hotswap  # Fast deploy (code only, ~15s)
make destroy         # Tear down the stack
make diff            # Show infrastructure changes
```

## 🔐 Authentication

The API supports two authentication methods:

### 1. Email/Password
```bash
POST /v1/auth/login
{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

### 2. Google OAuth
```bash
POST /v1/auth/google
{
  "id_token": "google-id-token..."
}

Response:
{
  "user_id": "uuid",
  "email": "user@gmail.com",
  "name": "User Name",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## 📚 Additional Documentation

- [Architecture Deep Dive](docs/AUTH_ARCHITECTURE.md)
- [Auth Module Refactoring](docs/AUTH_REFACTORING.md)
- [Google OAuth Setup](docs/GOOGLE_OAUTH_SETUP.md)

## 🧪 Testing

Unit tests use mocks for all external dependencies:

```bash
# Run all tests
make test

# Run tests with coverage
make cover

# Generate HTML coverage report
make cover-html
```

Example test structure:
```go
// Mock external dependencies
type mockUserRepo struct { mock.Mock }
type mockPasswordHasher struct { mock.Mock }

func TestLogin_Success(t *testing.T) {
    // Arrange
    repo := new(mockUserRepo)
    hasher := new(mockPasswordHasher)
    repo.On("FindByEmail", "user@example.com").Return(user, nil)
    hasher.On("Compare", hash, "password").Return(nil)

    // Act
    uc := NewLogin(repo, hasher, jwtGen)
    result, err := uc.Execute(ctx, LoginCommand{...})

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "token", result.Token)
}
```

## 🏷️ Tech Stack

- **Language**: Go 1.21+
- **Infrastructure**: AWS CDK (TypeScript)
- **Compute**: AWS Lambda (ARM64)
- **Database**: DynamoDB (Single Table Design)
- **API**: API Gateway REST API
- **Auth**: JWT tokens (HS256)
- **Testing**: testify/mock

## 📊 AWS Resources

| Resource | Name | Purpose |
|----------|------|---------|
| Lambda | `gym-tracker-auth` | Email/password + Google OAuth |
| Lambda | `gym-tracker-authorizer` | JWT validation |
| Lambda | `gym-tracker-user` | User CRUD |
| Lambda | `gym-tracker-plan` | Training plans |
| Lambda | `gym-tracker-session` | Workout tracking |
| DynamoDB | `gym-tracker` | Single table for all data |
| API Gateway | `gym-tracker-api` | REST API endpoints |
| SSM Parameter | `/gym-tracker/jwt-secret` | JWT signing key |
| SSM Parameter | `/gym-tracker/google-client-id` | Google OAuth client |

## 🤝 Contributing

When adding new features:

1. **Domain first**: Define entities and business rules
2. **Ports next**: Create interfaces for use cases and external dependencies
3. **Application**: Implement use cases with tests
4. **Adapters last**: Add HTTP handlers and infrastructure implementations

Follow the hexagonal architecture pattern to keep the codebase maintainable!

## 📝 License

[Add your license here]
