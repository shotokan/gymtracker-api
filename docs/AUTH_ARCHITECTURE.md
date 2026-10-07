# Auth Module Architecture

## Estructura Unificada

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT REQUEST                            │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              │                               │
      POST /auth/login              POST /auth/google
      { email, password }           { id_token }
              │                               │
              └───────────────┬───────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Lambda: gym-tracker-auth                       │
│                     (cmd/auth/main.go)                           │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ Router (adapter/primary/router.go)                      │    │
│  │  Switch on req.Path:                                    │    │
│  │  - /auth/login   → loginHandler.Handle()               │    │
│  │  - /auth/google  → googleAuthHandler.Handle()          │    │
│  └────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              │                               │
              ▼                               ▼
┌──────────────────────────┐   ┌──────────────────────────┐
│   LoginHandler           │   │  GoogleAuthHandler       │
│   (primary adapter)      │   │  (primary adapter)       │
└──────────┬───────────────┘   └───────────┬──────────────┘
           │                               │
           ▼                               ▼
┌──────────────────────────┐   ┌──────────────────────────┐
│   LoginUseCase           │   │  GoogleAuthUseCase       │
│   (application layer)    │   │  (application layer)     │
│                          │   │                          │
│  1. Find user by email   │   │  1. Validate Google token│
│  2. Compare password     │   │  2. Find/create user     │
│  3. Generate JWT         │   │  3. Link Google ID       │
└──────────┬───────────────┘   │  4. Generate JWT         │
           │                   └───────────┬──────────────┘
           │                               │
           └───────────────┬───────────────┘
                           │
                           ▼
         ┌─────────────────────────────────────┐
         │      Shared Dependencies             │
         │                                      │
         │  • UserRepository (DynamoDB)        │
         │  • JWTGenerator                     │
         │  • PasswordHasher (bcrypt)          │
         │  • GoogleTokenValidator             │
         └─────────────────────────────────────┘
```

## Flujo de Login (Email/Password)

```
1. POST /auth/login { email, password }
   │
2. LoginHandler.Handle()
   │
3. LoginUseCase.Execute(LoginCommand)
   │
   ├─→ UserRepository.FindByEmail(email)
   │   └─→ DynamoDB Query (PK = USER#email, SK = PROFILE)
   │
   ├─→ PasswordHasher.Compare(hash, password)
   │   └─→ bcrypt.CompareHashAndPassword()
   │
   └─→ JWTGenerator.Generate(userID, email)
       └─→ jwt.NewWithClaims(...)
   │
4. Return { token, expires_in }
```

## Flujo de Google OAuth

```
1. POST /auth/google { id_token }
   │
2. GoogleAuthHandler.Handle()
   │
3. GoogleAuthUseCase.Execute(GoogleAuthCommand)
   │
   ├─→ GoogleTokenValidator.ValidateToken(idToken)
   │   └─→ idtoken.Validate() [Google API]
   │   └─→ Extract { googleID, email, name, picture }
   │
   ├─→ UserRepository.FindByGoogleID(googleID)
   │   └─→ DynamoDB Query (GSI2: GSI2PK = GOOGLE#id, GSI2SK = PROFILE)
   │   └─→ Not found? Try FindByEmail()
   │       └─→ DynamoDB Query (PK = USER#email, SK = PROFILE)
   │
   ├─→ User exists?
   │   │
   │   ├─→ YES with GoogleID → Return user
   │   │
   │   ├─→ YES without GoogleID → Update user (link Google)
   │   │   └─→ UserRepository.Update(user with GoogleID)
   │   │
   │   └─→ NO → Create new user
   │       └─→ UserRepository.Save(new User {
   │           ID: uuid.New(),
   │           Email: googleUser.Email,
   │           GoogleID: googleUser.GoogleID,
   │           Name: googleUser.Name,
   │           PasswordHash: "" (empty for OAuth users)
   │       })
   │
   └─→ JWTGenerator.Generate(userID, email)
       └─→ jwt.NewWithClaims(...)
   │
4. Return { user_id, email, name, token }
```

## Componentes por Capa

### Domain Layer
```go
// internal/auth/domain/

// Errors
ErrInvalidCreds     // Login failed
ErrInvalidToken     // Google token invalid

// Models
GoogleUser {        // User info from Google
    GoogleID string
    Email    string
    Name     string
    Picture  string
}
```

### Application Layer (Use Cases)
```go
// internal/auth/application/

LoginUseCase {
    userRepo UserRepository
    hasher   PasswordHasher
    jwtGen   jwtGenerator
}

GoogleAuthUseCase {
    validator GoogleTokenValidator
    userRepo  UserRepository
    idGen     idGenerator
    jwtGen    jwtGenerator
}
```

### Adapters

**Primary (Handlers)**
```go
// internal/auth/adapter/primary/

LoginHandler {
    uc LoginUseCase
}

GoogleAuthHandler {
    uc GoogleAuthUseCase
}
```

**Secondary (Infrastructure)**
```go
// internal/auth/adapter/secondary/

bcrypt/PasswordHasher       // bcrypt.CompareHashAndPassword
google/TokenValidator       // idtoken.Validate (Google API)
jwt/TokenGenerator          // jwt.NewWithClaims
```

## DynamoDB Access Patterns

### Login (Email/Password)
```
Operation: FindByEmail
Query: GetItem
  PK = USER#<email>
  SK = PROFILE
```

### Google Auth - Find by Google ID
```
Operation: FindByGoogleID
Query: Query on GSI2
  GSI2PK = GOOGLE#<googleID>
  GSI2SK = PROFILE
```

### Google Auth - Find by Email (fallback)
```
Operation: FindByEmail
Query: GetItem
  PK = USER#<email>
  SK = PROFILE
```

### Google Auth - Create New User
```
Operation: Save
PutItem:
  PK = USER#<email>
  SK = PROFILE
  GSI1PK = USER_ID#<uuid>
  GSI1SK = PROFILE
  GSI2PK = GOOGLE#<googleID>  ← New for Google users
  GSI2SK = PROFILE
  google_id = <googleID>
  password_hash = ""           ← Empty for OAuth users
```

### Google Auth - Link Existing User
```
Operation: Update
UpdateItem:
  PK = USER#<email>
  SK = PROFILE
  SET google_id = <googleID>
  SET GSI2PK = GOOGLE#<googleID>
  SET GSI2SK = PROFILE
```

## JWT Token Structure

Ambos métodos de autenticación generan el **mismo formato** de JWT:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "iat": 1234567890,
  "exp": 1234654290,
  "iss": "gymtracker-api"
}
```

Firmado con `JWT_SECRET` (HS256)

## Environment Variables

```bash
# Required by cmd/auth/main.go
TABLE_NAME        # DynamoDB table name
JWT_SECRET        # Secret for signing JWT tokens
GOOGLE_CLIENT_ID  # OAuth 2.0 Client ID from Google Cloud Console
```

## Testing Strategy

### Unit Tests
```go
// Mocks for all dependencies
mockUserRepo          // DynamoDB operations
mockPasswordHasher    // bcrypt operations
mockJWTGen           // JWT generation
mockGoogleValidator  // Google API calls
mockIDGen            // UUID generation

// Test cases
✅ Login: success, user not found, invalid password
✅ Google: new user, existing user, link account, invalid token
```

### Integration Points
```
LoginUseCase ─────┬─→ UserRepository (shared with User module)
                  └─→ JWTGenerator (shared)

GoogleAuthUseCase ┬─→ UserRepository (shared with User module)
                  └─→ JWTGenerator (shared)

AuthorizerUseCase ──→ JWTGenerator (validate only)
```

## Dependencies Between Modules

```
auth module
  ├── Depends on: user/domain (User model)
  └── Used by: None (leaf module)

authorizer module
  ├── Depends on: auth/adapter/secondary/jwt (validation)
  └── Independent (only validates tokens)

user module
  └── Independent (no auth dependencies)
```

## Comparison: Before vs After

| Aspect | Before | After |
|--------|--------|-------|
| Lambda functions | `login`, `google-auth` | `auth` (unified) |
| Modules | `internal/login`, `internal/auth` | `internal/auth` |
| Code duplication | JWT generation in both | Single implementation |
| Routing | 2 separate Lambdas | 1 Lambda with path-based routing |
| Warm starts | Separate pools | Shared pool (faster) |
| Testing | Spread across modules | Unified test suite |
| Maintainability | Update 2 places | Update 1 place |
