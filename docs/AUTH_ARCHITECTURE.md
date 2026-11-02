# Auth Module Architecture

La API autentica únicamente con el correo y la contraseña de una cuenta registrada.
Las cuentas se crean con estado `pending` con `POST /users` en el módulo `user`, y un administrador
las aprueba con `PATCH /users/{id}/activate`. Solo las cuentas `active` pueden iniciar sesión.

## Estructura

```
POST /auth/login { email, password }
              │
              ▼
Lambda: gym-tracker-auth (cmd/auth/main.go)
              │
              ▼
Router (adapter/primary/router.go)
              │
              ▼
LoginHandler (adapter/primary/login_handler.go)
              │
              ▼
LoginUseCase (application/login_use_case.go)
              │
              ├─→ UserRepository (DynamoDB)
              ├─→ PasswordHasher (bcrypt)
              └─→ JWTGenerator
```

## Flujo de Login

```
1. POST /auth/login { email, password }
   │
2. LoginHandler.Handle()
   │
3. LoginUseCase.Execute(LoginCommand)
   │
   ├─→ UserRepository.FindByEmail(email)
   │   └─→ DynamoDB GetItem (PK = USER#email, SK = PROFILE)
   │
   ├─→ PasswordHasher.Compare(hash, password)
   │   └─→ bcrypt.CompareHashAndPassword()
   │
   └─→ JWTGenerator.Generate(userID, email)
       └─→ jwt.NewWithClaims(...)
   │
4. Return { token, expires_in }
```

Un correo inexistente, una contraseña incorrecta y una cuenta pendiente o dada de baja devuelven el mismo
`401 invalid credentials`
para evitar la enumeración de usuarios.

## Componentes por Capa

### Domain Layer
```go
// internal/auth/domain/

ErrInvalidCreds     // Login failed
```

### Application Layer (Use Cases)
```go
// internal/auth/application/

LoginUseCase {
    userRepo UserRepository
    hasher   PasswordHasher
    jwtGen   jwtGenerator
}
```

### Adapters

**Primary (Handlers)**
```go
// internal/auth/adapter/primary/

LoginHandler {
    uc LoginUseCase
}
```

**Secondary (Infrastructure)**
```go
// internal/auth/adapter/secondary/

bcrypt/PasswordHasher       // bcrypt.CompareHashAndPassword
jwt/TokenGenerator          // jwt.NewWithClaims
```

## DynamoDB Access Patterns

```
Operation: FindByEmail
Query: GetItem
  PK = USER#<email>
  SK = PROFILE
```

La Lambda de auth solo tiene permisos de lectura sobre la tabla.

## JWT Token Structure

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
```

## Testing Strategy

```go
// Mocks for all dependencies
mockUserRepo          // DynamoDB operations
mockPasswordHasher    // bcrypt operations
mockJWTGen            // JWT generation

// Test cases
✅ Login: success, user not found, invalid password, inactive user
```

## Dependencies Between Modules

```
auth module
  ├── Depends on: user/domain (User model)
  └── Used by: None (leaf module)

authorizer module
  └── Independent (only validates tokens)

user module
  └── Independent (no auth dependencies)
```
