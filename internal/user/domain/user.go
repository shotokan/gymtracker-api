package domain

import "time"

// Roles disponibles para un usuario.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Estados posibles de una cuenta. Solo StatusActive da acceso a la API.
const (
	StatusPending     = "pending"     // registrada, a la espera de aprobación
	StatusActive      = "active"      // aprobada por un administrador
	StatusDeactivated = "deactivated" // dada de baja
)

// User es la entidad central del dominio de usuarios.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Deactivate marca al usuario como inactivo (soft delete).
func (u *User) Deactivate() {
	u.Status = StatusDeactivated
	u.UpdatedAt = time.Now().UTC()
}

// Activate marca al usuario como activo, dándole acceso a la API.
func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now().UTC()
}

// IsActive indica si la cuenta tiene acceso a la API.
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsActiveAdmin indica si el usuario es un administrador con la cuenta activa.
func (u *User) IsActiveAdmin() bool {
	return u.IsActive() && u.Role == RoleAdmin
}

// UpdateName actualiza el nombre del usuario.
func (u *User) UpdateName(name string) {
	u.Name = name
	u.UpdatedAt = time.Now().UTC()
}
