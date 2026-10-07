package domain

import "time"

// User es la entidad central del dominio de usuarios.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	GoogleID     string // Google OAuth sub claim (empty if not using Google auth)
	Name         string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Deactivate marca al usuario como inactivo (soft delete).
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now().UTC()
}

// UpdateName actualiza el nombre del usuario.
func (u *User) UpdateName(name string) {
	u.Name = name
	u.UpdatedAt = time.Now().UTC()
}
